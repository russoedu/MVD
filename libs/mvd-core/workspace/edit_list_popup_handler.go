package workspace

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// editSavedMsg is sent when the person saves the edit of the list.
type editSavedMsg struct{ items []playlistfile.Item }

// editCancelledMsg is sent when the person leaves the edit without saving.
type editCancelledMsg struct{}

// editPopup is the list as it can be edited: each address, and under it its songs by
// name. Addresses and songs can be added to, removed and moved; nothing reaches the list
// until the person saves.
type editPopup struct {
	items []playlistfile.Item
	sel   int // the row, counting each address and each song of it
	input textinput.Model
	// asking is "" for none, "add" for a new address, "edit" for the address of the row.
	asking string
}

// editRow is one line of the popup: an address (song is -1) or one of its songs.
type editRow struct{ item, song int }

func newEditPopup(file playlistfile.File) editPopup {
	items := make([]playlistfile.Item, len(file.Items))
	for i, it := range file.Items {
		items[i] = playlistfile.Item{Source: it.Source, Songs: append([]playlistfile.Entry(nil), it.Songs...)}
	}
	return editPopup{items: items}
}

func (p editPopup) rows() []editRow {
	var rows []editRow
	for i, it := range p.items {
		rows = append(rows, editRow{i, -1})
		for j := range it.Songs {
			rows = append(rows, editRow{i, j})
		}
	}
	return rows
}

func (p editPopup) current() (editRow, bool) {
	rows := p.rows()
	if p.sel < 0 || p.sel >= len(rows) {
		return editRow{}, false
	}
	return rows[p.sel], true
}

// selectRow puts the selection on the given row.
func (p editPopup) selectRow(r editRow) editPopup {
	for i, row := range p.rows() {
		if row == r {
			p.sel = i
		}
	}
	return p
}

func (p editPopup) Update(msg tea.KeyPressMsg) (editPopup, tea.Cmd) {
	if p.asking != "" {
		return p.answer(msg)
	}
	switch msg.String() {
	case "esc":
		return p, func() tea.Msg { return editCancelledMsg{} }
	case "s", "ctrl+s":
		items := p.items
		return p, func() tea.Msg { return editSavedMsg{items: items} }
	case "up":
		p.sel = max(p.sel-1, 0)
	case "down":
		p.sel = min(p.sel+1, max(len(p.rows())-1, 0))
	case "a":
		return p.ask("add", "Address to add", ""), textinput.Blink
	case "enter", "e":
		if r, ok := p.current(); ok && r.song < 0 {
			return p.ask("edit", "Address", p.items[r.item].Source), textinput.Blink
		}
	case "d", "delete":
		return p.remove(), nil
	case "K", "shift+up":
		return p.move(-1), nil
	case "J", "shift+down":
		return p.move(1), nil
	}
	return p, nil
}

func (p editPopup) ask(kind, prompt, value string) editPopup {
	in := textinput.New()
	in.Prompt = prompt + ": "
	in.SetValue(value)
	in.CursorEnd()
	in.Focus()
	p.asking, p.input = kind, in
	return p
}

func (p editPopup) answer(msg tea.KeyPressMsg) (editPopup, tea.Cmd) {
	switch msg.String() {
	case "esc":
		p.asking = ""
		return p, nil
	case "enter":
		text := strings.TrimSpace(p.input.Value())
		kind := p.asking
		p.asking = ""
		if text == "" {
			return p, nil
		}
		if kind == "add" {
			at := len(p.items)
			if r, ok := p.current(); ok {
				at = r.item + 1
			}
			items := append([]playlistfile.Item(nil), p.items[:at]...)
			items = append(items, playlistfile.Item{Source: text})
			p.items = append(items, p.items[at:]...)
			return p.selectRow(editRow{at, -1}), nil
		}
		if r, ok := p.current(); ok && p.items[r.item].Source != text {
			// An address that changed is another list of songs: the ones it had are no longer its own.
			p.items[r.item] = playlistfile.Item{Source: text}
			return p.selectRow(editRow{r.item, -1}), nil
		}
		return p, nil
	}
	in, cmd := p.input.Update(msg)
	p.input = in
	return p, cmd
}

// remove takes out the song or the address (with its songs) the selection is on.
func (p editPopup) remove() editPopup {
	r, ok := p.current()
	if !ok {
		return p
	}
	if r.song < 0 {
		p.items = append(append([]playlistfile.Item(nil), p.items[:r.item]...), p.items[r.item+1:]...)
	} else {
		songs := p.items[r.item].Songs
		p.items[r.item].Songs = append(append([]playlistfile.Entry(nil), songs[:r.song]...), songs[r.song+1:]...)
	}
	p.sel = min(p.sel, max(len(p.rows())-1, 0))
	return p
}

// move shifts the song or the address the selection is on one place up or down among its
// own kind, and the selection goes with it.
func (p editPopup) move(by int) editPopup {
	r, ok := p.current()
	if !ok {
		return p
	}
	if r.song < 0 {
		to := r.item + by
		if to < 0 || to >= len(p.items) {
			return p
		}
		p.items[r.item], p.items[to] = p.items[to], p.items[r.item]
		return p.selectRow(editRow{to, -1})
	}
	songs := p.items[r.item].Songs
	to := r.song + by
	if to < 0 || to >= len(songs) {
		return p
	}
	songs[r.song], songs[to] = songs[to], songs[r.song]
	return p.selectRow(editRow{r.item, to})
}

// View draws the popup w by h cells.
func (p editPopup) View(w, h int) string {
	inner := max(w-2, 1)
	room := max(h-2-2, 1) // the frame, and the line of keys and the question
	rows := p.rows()
	top := 0
	if p.sel >= room {
		top = p.sel - room + 1
	}
	var lines []string
	for i := top; i < min(len(rows), top+room); i++ {
		r := rows[i]
		var text string
		if r.song < 0 {
			text = "  " + cells.Display(p.items[r.item].Source)
		} else {
			text = "      " + songName(p.items[r.item].Songs[r.song])
		}
		if i == p.sel {
			lines = append(lines, look.Key.Render(cells.Fit("▸"+text[1:], inner)))
		} else if r.song >= 0 {
			lines = append(lines, look.Dim.Render(cells.Fit(text, inner)))
		} else {
			lines = append(lines, cells.Fit(text, inner))
		}
	}
	if len(rows) == 0 {
		lines = append(lines, look.Dim.Render("  The list is empty: press a to add an address"))
	}
	for len(lines) < room {
		lines = append(lines, "")
	}
	if p.asking != "" {
		lines = append(lines, look.Key.Render(" "+p.input.View()))
	} else {
		lines = append(lines, look.Dim.Render(" ↑↓ move  a add  ⏎ change address  d remove  K/J move up/down  s save  esc cancel"))
	}
	return dialog.Frame(fmt.Sprintf("Edit the list · %d addresses", len(p.items)), lines, w, h)
}

// songName is what a song is called: the artist and the title a music database gave it,
// else the upload's own title.
func songName(s playlistfile.Entry) string {
	name := cells.Display(s.Upload.Title)
	if s.Artist != "" && s.Title != "" {
		name = cells.Display(s.Artist + " - " + s.Title)
	}
	if name == "" {
		name = "(" + s.Upload.ID + ")"
	}
	return name
}
