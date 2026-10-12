package download

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/connect"
	addlinkspage "youtube-downloader/libs/mvd-core/run/addlinks"
	helppage "youtube-downloader/libs/mvd-core/run/help"
	"youtube-downloader/libs/mvd-core/keybar"
)

// page is the name of this screen in the design, as the designer knows it.
const page = "Download"

// liveRows is the most rows of a pane the connection publishes: the designer draws what it
// has room for, and a full log would only slow the line down.
const liveRows = 500

type (
	// designerMsg is an event the designer raised by a click or key on its drawing.
	designerMsg connect.Event
	// designerClosed says the connection ended.
	designerClosed struct{}
)

func waitDesigner(l *connect.Link) tea.Cmd {
	if l == nil {
		return nil
	}
	return func() tea.Msg {
		ev, ok := l.Next()
		if !ok {
			return designerClosed{}
		}
		return designerMsg(ev)
	}
}

// Live is the page the screen shows now and its inputs, by the input names of its design,
// for the live connection to the designer: the help and the box for adding links are pages
// of their own while they are open, and a pane gives every row, not the rows that fit.
func (m Model) Live() (pageName string, inputs map[string]any) {
	w, h := m.width, m.height
	switch {
	case m.showHelp:
		return "Help", helppage.Live(w, h, m.ctrl.LogPath())
	case m.adding:
		props := addlinkspage.Default()
		props.Hints = keybar.Of[addlinkspage.HintsHint](addLinksHints())
		props.LinksView = m.links.View()
		return "AddLinks", connect.Plain(addlinkspage.Inputs(props))
	}
	p := m.downloadProps(w, h)
	p.Playlists, _ = p.PlaylistsRows(w, liveRows)
	p.Entries, _ = p.EntriesRows(w, liveRows)
	p.Notes, _ = p.NotesRows(w, liveRows)
	p.Output, _ = p.OutputRows(w, liveRows)
	return page, connect.Plain(Inputs(p))
}

// publish tells the designer what the screen shows, when it is connected and is the
// screen that listens: inside the app the router does it for every screen.
func (m Model) publish() {
	if m.link == nil || m.width == 0 || m.height == 0 {
		return
	}
	name, inputs := m.Live()
	m.link.Publish(name, inputs, true)
}

// designerEvent does what a click on the designer's drawing does on the screen itself.
func (m Model) designerEvent(msg designerMsg) (tea.Model, tea.Cmd) {
	if msg.Page != page {
		return m, waitDesigner(m.link)
	}
	next, cmd := m.ApplyDesigner(msg.Name, msg.X, msg.Y)
	return next, tea.Batch(cmd, waitDesigner(m.link))
}

// ApplyDesigner does what a click on the designer's drawing does: a key bar entry presses
// its key, and a playlist or an entry is selected. x and y are the cell inside the
// component that was clicked.
func (m Model) ApplyDesigner(name string, x, y int) (Model, tea.Cmd) {
	if m.showHelp || m.confirmQuit || m.adding {
		return m, nil
	}
	ev, ok := EventNamed(name, x, y)
	if !ok {
		return m, nil
	}
	frame := Layout(m.downloadProps(m.width, m.height), m.width, m.height)
	switch ev.(type) {
	case HintEvent:
		if h, found := keybar.At(m.hints(), x); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				next, cmd := m.handleKey(key)
				return next.(Model), cmd
			}
		}
	case PickPlaylistEvent:
		if r, found := frame.RegionNamed("Playlists rows"); found {
			m.clickPlaylist(r.Y + y)
		}
	case PickEntryEvent:
		if r, found := frame.RegionNamed("Entries rows"); found {
			m.clickEntry(r.Y + y)
		}
	}
	return m, nil
}
