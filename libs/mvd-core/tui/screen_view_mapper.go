package tui

import tea "charm.land/bubbletea/v2"

// screenView wraps a rendered screen as the full screen view every model returns:
// on the alternate screen, with the mouse reported for clicks and the wheel.
func screenView(content string) tea.View {
	view := tea.NewView(content)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}
