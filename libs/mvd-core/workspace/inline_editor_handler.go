package workspace

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/lineedit"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/pointer"
)

// gripCells are the columns at the left of an address being edited: the grip and a space.
const gripCells = 2

// doubleClickWithin is how soon a second click makes a double click.
const doubleClickWithin = 450 * time.Millisecond

// inlineEdit is an address being written in the row it belongs to: a new one, in the row it
// will take, or one that is being changed.
type inlineEdit struct {
	on     bool
	adding bool
	index  int // the row, counted in the list as it will be
	// before is the playlist the cursor was on, for a cancelled add to put it back.
	before int
	line   lineedit.Model
}

// lastClick is the last left click, to tell a double click.
type lastClick struct {
	x, y int
	at   time.Time
}

// isDouble reports whether a click at (x, y) now is the second of a double click.
func (m Model) isDouble(x, y int, now time.Time) bool {
	c := m.clicked
	return !c.at.IsZero() && now.Sub(c.at) <= doubleClickWithin && c.y == y && abs(c.x-x) <= 1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// startAdd opens a new row after the playlist the cursor is on, to write an address in.
func (m Model) startAdd() (Model, tea.Cmd) {
	if !m.editable() {
		return m, nil
	}
	m = m.stopInline(false)
	items := cloneItems(m.session.File().Items)
	at, before := len(items), len(items)-1
	if i, ok := m.selected(); ok {
		at, before = i+1, i
	}
	shown := append(append(append([]playlistfile.Item(nil), items[:at]...), playlistfile.Item{}), items[at:]...)
	m.screen = m.screen.WithState(downloadlist.PreviewOf(playlistfile.File{Items: shown})).SelectPlaylist(at)
	m.inline = inlineEdit{on: true, adding: true, index: at, before: before}
	return m.sizeInline(""), nil
}

// startChange turns the row of playlist i into a place to write another address.
func (m Model) startChange(i int) (Model, tea.Cmd) {
	items := m.session.File().Items
	if !m.editable() || i < 0 || i >= len(items) {
		return m, nil
	}
	m = m.stopInline(false)
	m.screen = m.screen.SelectPlaylist(i)
	m.inline = inlineEdit{on: true, index: i, before: i}
	return m.sizeInline(items[i].Source), nil
}

// sizeInline makes the editor for text as wide as its row.
func (m Model) sizeInline(text string) Model {
	width := 20
	if _, _, w, ok := m.screen.PlaylistRow(m.inline.index); ok {
		width = max(w-gripCells, 1)
	}
	m.inline.line = lineedit.New(text, width)
	return m
}

// stopInline ends the editing: with keep, what was written becomes the address; without, the
// list is as it was.
func (m Model) stopInline(keep bool) Model {
	if !m.inline.on {
		return m
	}
	e := m.inline
	m.inline = inlineEdit{}
	text := strings.TrimSpace(e.line.Value())
	switch {
	case keep && text != "" && e.adding:
		links, _ := parseLinks(text)
		added := make([]playlistfile.Item, 0, len(links))
		for _, link := range links {
			added = append(added, playlistfile.Item{Source: link})
		}
		items := cloneItems(m.session.File().Items)
		out := append(append(append([]playlistfile.Item(nil), items[:e.index]...), added...), items[e.index:]...)
		return m.setItems(out, e.index)
	case keep && text != "":
		return m.changeAddress(e.index, text)
	}
	m.screen = m.screen.WithState(m.session.Preview()).SelectPlaylist(e.before)
	return m
}

// inlineRow is where the row being written is, on the screen.
func (m Model) inlineRow() (x, y, w int, ok bool) {
	x, y, w, ok = m.screen.PlaylistRow(m.inline.index)
	return x, y + topRows, w, ok
}

// inlineView lays the editor over its row.
func (m Model) inlineView(screen string) string {
	x, y, w, ok := m.inlineRow()
	if !ok {
		return screen
	}
	m.inline.line = m.inline.line.SetWidth(max(w-gripCells, 1))
	screen = dialog.Overlay(screen, "≡ "+m.inline.line.View(), x, y)
	if m.inline.line.Value() == "" && w-gripCells > 2 {
		// An empty row says what it is for, after the cursor.
		hint := cells.Truncate("Paste a link here", w-gripCells-1)
		screen = dialog.Overlay(screen, lipgloss.NewStyle().Background(lipgloss.Color(look.Colours.Selected)).Foreground(lipgloss.Color(look.Colours.Dim)).Render(hint), x+gripCells+1, y)
	}
	return screen
}

// inlineKey gives a key to the editor: enter keeps what was written, esc drops it.
func (m Model) inlineKey(k tea.KeyPressMsg) (Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		return m.stopInline(false), nil
	case "enter":
		if reason := notALink(m.inline.line.Value()); reason != "" {
			m.notice = reason // the row stays open: the person fixes it where it is
			return m, nil
		}
		return m.stopInline(true), nil
	case "ctrl+c", "ctrl+x", "ctrl+v":
		if m.host.Clipboard != nil {
			name := map[string]string{"ctrl+c": entryCopy, "ctrl+x": entryCut, "ctrl+v": entryPaste}[k.String()]
			next, cmd := m.editCommand(name)
			return next.(Model), cmd
		}
	}
	line, cmd := m.inline.line.Update(k)
	m.inline.line = line
	return m, cmd
}

// inlineMouse gives the mouse to the editor. A click outside its row keeps what was written
// and is then worked as it would have been.
func (m Model) inlineMouse(msg tea.MouseMsg, double bool) (Model, tea.Cmd, bool) {
	x, y, w, ok := m.inlineRow()
	pos := msg.Mouse()
	inside := ok && pos.Y == y && pos.X >= x+gripCells && pos.X < x+w
	switch msg.(type) {
	case tea.MouseMotionMsg, tea.MouseReleaseMsg:
		line, cmd := m.inline.line.Update(pointer.Shift(msg, -(x + gripCells), 0))
		m.inline.line = line
		return m, cmd, true
	case tea.MouseClickMsg:
		if inside {
			line, cmd := m.inline.line.Update(pointer.Shift(msg, -(x + gripCells), 0))
			m.inline.line = line
			if double {
				m.inline.line = m.inline.line.SelectAll()
			}
			return m, cmd, true
		}
		if pos.Button == tea.MouseLeft && pos.Y != 0 {
			m = m.stopInline(true)
		}
	}
	return m, nil, false
}

// pasteInline puts the first line of text in the editor.
func (m Model) pasteInline(text string) Model {
	line := strings.TrimSpace(strings.SplitN(strings.TrimSpace(text), "\n", 2)[0])
	m.inline.line = m.inline.line.Paste(line)
	return m
}

// listMouse works the mouse on the list before the screen sees it: the editor of the row
// being written takes its clicks and its dragging, and a double click on the playlists
// changes the playlist under it, or adds one when there is none. It gives back the model
// whether or not it handled the event, as it keeps the last click.
func (m Model) listMouse(msg tea.MouseMsg) (Model, tea.Cmd, bool) {
	pos := msg.Mouse()
	click, isClick := msg.(tea.MouseClickMsg)
	double := false
	if isClick && click.Button == tea.MouseLeft {
		now := time.Now()
		double = m.isDouble(click.X, click.Y, now)
		m.clicked = lastClick{x: click.X, y: click.Y, at: now}
		if double {
			m.clicked = lastClick{} // a third click is not another double
		}
	}
	if m.inline.on {
		next, cmd, handled := m.inlineMouse(msg, double)
		if handled {
			return next, cmd, true
		}
		m = next
	}
	if !double || m.reviewing || !m.editable() || pos.Y == 0 {
		return m, nil, false
	}
	index, inside := m.screen.PlaylistAt(pos.X, pos.Y-topRows)
	switch {
	case !inside:
		return m, nil, false
	case index >= 0:
		next, cmd := m.startChange(index)
		return next, cmd, true
	}
	next, cmd := m.startAdd()
	return next, cmd, true
}
