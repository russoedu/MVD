package workspace

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/playlistfile"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
)

// editable reports whether the list can be changed now: not while a run has its own copy.
func (m Model) editable() bool { return !m.running() && !m.starting && !m.stopping }

func cloneItems(items []playlistfile.Item) []playlistfile.Item {
	out := make([]playlistfile.Item, len(items))
	for i, it := range items {
		out[i] = playlistfile.Item{Source: it.Source, Songs: append([]playlistfile.Entry(nil), it.Songs...)}
	}
	return out
}

// setItems makes items the list, keeping what it was to undo, and puts the cursor on playlist
// sel.
func (m Model) setItems(items []playlistfile.Item, sel int) Model {
	m.undone = append(m.undone, cloneItems(m.session.File().Items))
	m.redone = nil
	return m.showItems(items, sel)
}

// showItems writes items as the list and the screen shows it, without keeping the old one.
func (m Model) showItems(items []playlistfile.Item, sel int) Model {
	if err := m.session.Replace(items); err != nil {
		m.notice = "Cannot save the list: " + err.Error()
	}
	m.screen = m.screen.WithState(m.session.Preview()).SelectPlaylist(sel)
	return m
}

// selected is the playlist the cursor is on, as an index of the list, and whether there is one.
func (m Model) selected() (int, bool) {
	i := m.screen.SelectedPlaylist()
	return i, i >= 0 && i < len(m.session.File().Items)
}

// editList does what the person did to a playlist of the screen.
func (m Model) editList(msg downloadpage.ListEditMsg) (tea.Model, tea.Cmd) {
	if !m.editable() {
		m.notice = "Stop the downloads before editing the list"
		return m, nil
	}
	items := cloneItems(m.session.File().Items)
	if msg.Index < 0 || msg.Index >= len(items) {
		return m, nil
	}
	switch msg.Op {
	case "remove":
		m = m.setItems(append(items[:msg.Index], items[msg.Index+1:]...), msg.Index)
	case "move":
		if msg.To < 0 || msg.To >= len(items) || msg.To == msg.Index {
			return m, nil
		}
		moved := items[msg.Index]
		rest := append(append([]playlistfile.Item(nil), items[:msg.Index]...), items[msg.Index+1:]...)
		out := append(append([]playlistfile.Item(nil), rest[:msg.To]...), moved)
		m = m.setItems(append(out, rest[msg.To:]...), msg.To)
	case "edit":
		return m.startChange(msg.Index)
	}
	return m, nil
}

// changeAddress gives playlist i another address; the songs it had are no longer its own.
func (m Model) changeAddress(i int, address string) Model {
	items := cloneItems(m.session.File().Items)
	if i < 0 || i >= len(items) || items[i].Source == address {
		return m
	}
	items[i] = playlistfile.Item{Source: address}
	return m.setItems(items, i)
}

// addAddresses adds one playlist for each non-empty line of text, after the one the cursor
// is on (or at the end), and puts the cursor on the first one added.
func (m Model) addAddresses(text string) Model {
	if !m.editable() {
		return m
	}
	var added []playlistfile.Item
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			added = append(added, playlistfile.Item{Source: line})
		}
	}
	if len(added) == 0 {
		return m
	}
	items := cloneItems(m.session.File().Items)
	at := len(items)
	if i, ok := m.selected(); ok {
		at = i + 1
	}
	out := append(append([]playlistfile.Item(nil), items[:at]...), added...)
	return m.setItems(append(out, items[at:]...), at)
}

// editCommand works the commands of the Edit menu on the playlist the cursor is on.
func (m Model) editCommand(name string) (tea.Model, tea.Cmd) {
	if !m.editable() {
		m.notice = "Stop the downloads before editing the list"
		return m, nil
	}
	i, ok := m.selected()
	switch name {
	case entryUndo:
		if len(m.undone) == 0 {
			return m, nil
		}
		m.redone = append(m.redone, cloneItems(m.session.File().Items))
		last := m.undone[len(m.undone)-1]
		m.undone = m.undone[:len(m.undone)-1]
		return m.showItems(last, i), nil
	case entryRedo:
		if len(m.redone) == 0 {
			return m, nil
		}
		m.undone = append(m.undone, cloneItems(m.session.File().Items))
		last := m.redone[len(m.redone)-1]
		m.redone = m.redone[:len(m.redone)-1]
		return m.showItems(last, i), nil
	case entryPaste:
		return m.paste()
	case entryCopy:
		return m.copy(false)
	case entryCut:
		return m.copy(true)
	}
	if !ok {
		return m, nil
	}
	switch name {
	case entryDelete:
		return m.editList(downloadpage.ListEditMsg{Op: "remove", Index: i})
	case entryMoveUp:
		return m.editList(downloadpage.ListEditMsg{Op: "move", Index: i, To: i - 1})
	case entryMoveDown:
		return m.editList(downloadpage.ListEditMsg{Op: "move", Index: i, To: i + 1})
	}
	return m, nil
}

// copy puts what is selected in the address being written on the clipboard; with cut it is taken
// out too.
func (m Model) copy(cut bool) (tea.Model, tea.Cmd) {
	if !m.inline.on || !m.inline.line.HasSelection() {
		return m, nil
	}
	text := m.inline.line.Selection()
	if cut {
		m.inline.line, _ = m.inline.line.Cut()
	}
	if m.host.Clipboard != nil {
		if err := m.host.Clipboard.Write(text); err != nil {
			m.notice = "Cannot copy: " + err.Error()
		}
		return m, nil
	}
	return m, tea.SetClipboard(text)
}

// paste puts the clipboard in the address being written, or adds its links to the list.
func (m Model) paste() (tea.Model, tea.Cmd) {
	if m.host.Clipboard == nil {
		return m, tea.ReadClipboard // the terminal answers with a ClipboardMsg
	}
	text, err := m.host.Clipboard.Read()
	if err != nil {
		m.notice = "Cannot paste: " + err.Error()
		return m, nil
	}
	return m.pasted(text), nil
}

// pasted takes text that came from the clipboard.
func (m Model) pasted(text string) Model {
	if m.inline.on {
		return m.pasteInline(text)
	}
	return m.addAddresses(text)
}
