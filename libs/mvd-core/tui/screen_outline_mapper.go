package tui

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
