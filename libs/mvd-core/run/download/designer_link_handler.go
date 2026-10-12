package download

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/connect"
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

// publish tells the designer what the screen shows, by the input names of its design: the
// values the screen gives its page, with every row of a pane in place of the rows that fit.
func (m Model) publish() {
	if m.link == nil || m.width == 0 || m.height == 0 {
		return
	}
	w, h := m.width, m.height
	p := m.downloadProps(w, h)
	p.Playlists, _ = p.PlaylistsRows(w, liveRows)
	p.Entries, _ = p.EntriesRows(w, liveRows)
	p.Notes, _ = p.NotesRows(w, liveRows)
	p.Output, _ = p.OutputRows(w, liveRows)
	inputs := Inputs(p)
	for name, v := range inputs {
		// The screen colours some of its texts with escape codes; the designer draws its own.
		if text, ok := v.(string); ok {
			inputs[name] = ansi.Strip(text)
		}
	}
	m.link.Publish(page, inputs, true)
}

// designerEvent does what a click on the designer's drawing does on the screen itself.
func (m Model) designerEvent(msg designerMsg) (tea.Model, tea.Cmd) {
	if msg.Page != page || m.showHelp || m.confirmQuit || m.adding {
		return m, waitDesigner(m.link)
	}
	ev, ok := EventNamed(msg.Name, msg.X, msg.Y)
	if !ok {
		return m, waitDesigner(m.link)
	}
	frame := Layout(m.downloadProps(m.width, m.height), m.width, m.height)
	switch ev.(type) {
	case HintEvent:
		if h, found := keybar.At(m.hints(), msg.X); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				next, cmd := m.handleKey(key)
				return next, tea.Batch(cmd, waitDesigner(m.link))
			}
		}
	case PickPlaylistEvent:
		if r, found := frame.RegionNamed("Playlists rows"); found {
			m.clickPlaylist(r.Y + msg.Y)
		}
	case PickEntryEvent:
		if r, found := frame.RegionNamed("Entries rows"); found {
			m.clickEntry(r.Y + msg.Y)
		}
	}
	return m, waitDesigner(m.link)
}
