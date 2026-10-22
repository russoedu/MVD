package workspace

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/cells"
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
	// undone and redone are the states of the items before and after the changes made,
	// most recent last.
	undone, redone [][]playlistfile.Item
	// asking is "" for none, "add" for a new address, "edit" for the address of the row.
	asking string

	// w and h are the size of the popup, frame included; top is the first row that shows.
	w, h, top int
	// dragging is true while a row is being carried with the mouse; dropAt is the row it is over.
	dragging bool
	dropAt   int
}

// editRow is one line of the popup: an address (song is -1) or one of its songs.
type editRow struct{ item, song int }

func cloneItems(items []playlistfile.Item) []playlistfile.Item {
	out := make([]playlistfile.Item, len(items))
	for i, it := range items {
		out[i] = playlistfile.Item{Source: it.Source, Songs: append([]playlistfile.Entry(nil), it.Songs...)}
	}
	return out
}

// remember keeps the items as they are before a change, so it can be undone.
func (p editPopup) remember() editPopup {
	p.undone = append(p.undone, cloneItems(p.items))
	p.redone = nil
	return p
}

// undo goes back one change; redo goes forward again.
func (p editPopup) undo() editPopup {
	if len(p.undone) == 0 {
		return p
	}
	p.redone = append(p.redone, cloneItems(p.items))
	p.items, p.undone = p.undone[len(p.undone)-1], p.undone[:len(p.undone)-1]
	p.sel = min(p.sel, max(len(p.rows())-1, 0))
	return p
}

func (p editPopup) redo() editPopup {
	if len(p.redone) == 0 {
		return p
	}
	p.undone = append(p.undone, cloneItems(p.items))
	p.items, p.redone = p.redone[len(p.redone)-1], p.redone[:len(p.redone)-1]
	p.sel = min(p.sel, max(len(p.rows())-1, 0))
	return p
}

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

// Update takes a key and keeps the selection in view.
func (p editPopup) Update(msg tea.KeyPressMsg) (editPopup, tea.Cmd) {
	p, cmd := p.key(msg)
	return p.reveal(), cmd
}

func (p editPopup) key(msg tea.KeyPressMsg) (editPopup, tea.Cmd) {
	if p.asking != "" {
		return p.answer(msg)
	}
	switch msg.String() {
	case "esc":
		return p, func() tea.Msg { return editCancelledMsg{} }
	case "ctrl+s":
		items := p.items
		return p, func() tea.Msg { return editSavedMsg{items: items} }
	case "ctrl+z":
		return p.undo(), nil
	case "ctrl+y":
		return p.redo(), nil
	case "up":
		p.sel = max(p.sel-1, 0)
	case "down":
		p.sel = min(p.sel+1, max(len(p.rows())-1, 0))
	case "enter":
		if r, ok := p.current(); ok && r.song < 0 {
			return p.ask("edit", "Address", p.items[r.item].Source), textinput.Blink
		}
	case "ctrl+c":
		return p, p.copy()
	case "ctrl+v":
		return p, tea.ReadClipboard
	case "delete":
		return p.remove(), nil
	case "shift+up":
		return p.move(-1), nil
	case "shift+down":
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
			p = p.remember()
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
			p = p.remember()
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

// addAddresses adds one address for each non-empty line of text, after the address the
// selection is on (or at the end), and selects the first one added.
func (p editPopup) addAddresses(text string) editPopup {
	var added []playlistfile.Item
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			added = append(added, playlistfile.Item{Source: line})
		}
	}
	if len(added) == 0 {
		return p
	}
	p = p.remember()
	at := len(p.items)
	if r, ok := p.current(); ok {
		at = r.item + 1
	}
	items := append([]playlistfile.Item(nil), p.items[:at]...)
	items = append(items, added...)
	p.items = append(items, p.items[at:]...)
	return p.selectRow(editRow{at, -1})
}

// selectedText is what Copy puts on the clipboard: the address, or the name of the song,
// the selection is on.
func (p editPopup) selectedText() string {
	r, ok := p.current()
	if !ok {
		return ""
	}
	if r.song < 0 {
		return p.items[r.item].Source
	}
	return songName(p.items[r.item].Songs[r.song])
}

// copy puts the selected address or song on the clipboard.
func (p editPopup) copy() tea.Cmd {
	if text := p.selectedText(); text != "" {
		return tea.SetClipboard(text)
	}
	return nil
}

// remove takes out the song or the address (with its songs) the selection is on.
func (p editPopup) remove() editPopup {
	r, ok := p.current()
	if !ok {
		return p
	}
	p = p.remember()
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
		p = p.remember()
		p.items[r.item], p.items[to] = p.items[to], p.items[r.item]
		return p.selectRow(editRow{to, -1})
	}
	songs := p.items[r.item].Songs
	to := r.song + by
	if to < 0 || to >= len(songs) {
		return p
	}
	p = p.remember()
	songs = p.items[r.item].Songs
	songs[r.song], songs[to] = songs[to], songs[r.song]
	return p.selectRow(editRow{r.item, to})
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
