package tui

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/connect"
	"youtube-downloader/libs/mvd-core/keybar"
)

type (
	// designerMsg is an event the designer raised by a click or key on its drawing.
	designerMsg connect.Event
	// designerClosed says the live connection ended.
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

// live is the page the app shows now and its inputs, by the input names of the design.
func (m appModel) live() (page string, inputs map[string]any) {
	switch {
	case m.editing && !m.downloading:
		return m.editor.Live()
	case m.downloading:
		return m.download.Live()
	}
	return m.setup.live()
}

// publish tells the designer what the app shows, when it is connected.
func (m appModel) publish() {
	if m.link == nil || m.width == 0 || m.height == 0 {
		return
	}
	page, inputs := m.live()
	m.link.Publish(page, inputs, true)
}

// designerEvent does what a click on the designer's drawing does on the screen showing.
func (m appModel) designerEvent(ev connect.Event) (tea.Model, tea.Cmd) {
	wait := waitDesigner(m.link)
	page, _ := m.live()
	if ev.Page != page {
		return m, wait
	}
	switch {
	case m.editing && !m.downloading:
		next, cmd := m.editor.ApplyDesigner(ev.Name, ev.X, ev.Y)
		m.editor = next
		return m, tea.Batch(cmd, wait)
	case m.downloading:
		next, cmd := m.download.ApplyDesigner(ev.Name, ev.X, ev.Y)
		m.download = next
		return m, tea.Batch(cmd, wait)
	}
	next, cmd := m.setup.applyDesigner(ev)
	m.setup = next
	if cmd == nil {
		return m, wait
	}
	// A key on the key bar goes the way a key of the keyboard goes, through the app.
	return m, tea.Batch(cmd, wait)
}

// live is the page of the setup screen showing and its inputs.
func (m setupModel) live() (page string, inputs map[string]any) {
	switch m.screen {
	case screenConfig:
		return m.config.Live(m.width, m.height)
	case screenAdvanced:
		return m.advanced.Live(m.width, m.height)
	case screenColours:
		return m.colours.Live(m.width, m.height)
	case screenSongs:
		return m.songs.Live(m.width, m.height)
	}
	return m.list.Live(m.width, m.height)
}

// applyDesigner presses the key of a key bar entry the designer was clicked on, or selects
// the setting of a row. The key goes back as a message, so it takes the same way as one
// typed.
func (m setupModel) applyDesigner(ev connect.Event) (setupModel, tea.Cmd) {
	switch ev.Name {
	case "Hint":
		if h, found := keybar.At(m.hints(), ev.X); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				return m, func() tea.Msg { return key }
			}
		}
	case "Pick":
		// The title takes the first row of the screen.
		m.selectSetting(ev.Y + 1)
	}
	return m, nil
}
