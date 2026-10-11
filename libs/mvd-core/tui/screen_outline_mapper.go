package tui

import ()

// Outline describes the current screen. It reads the same state and the same
// key hints the views draw from, so the two cannot drift apart.
func (m setupModel) Outline() ScreenOutline {
	switch m.screen {
	case screenConfig:
		return m.config.Outline()
	case screenAdvanced:
		return m.advanced.Outline()
	case screenColours:
		return m.colours.Outline()
	case screenSongs:
		return m.songs.Outline()
	default:
		return m.list.Outline()
	}
}

// Outline describes the app's current screen: a setup screen, or the download
// screen.
func (m appModel) Outline() ScreenOutline {
	var out ScreenOutline
	switch {
	case m.stopping:
		return ScreenOutline{Title: "MVD · Stopping the downloads"}
	case m.editing && !m.downloading:
		return m.editor.Outline()
	case m.downloading:
		return m.download.Outline()
	default:
		out = m.setup.Outline()
	}
	if out.Prompt == "" {
		out.Prompt = m.notice
	}
	return out
}
