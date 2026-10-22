// Package screenview wraps what a screen drew as the view a Bubble Tea program shows.
package screenview

import (
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"

	tea "charm.land/bubbletea/v2"
)

// Of wraps a rendered screen as the full screen view every model returns:
// on the alternate screen, with the mouse reported for clicks and the wheel.
func Of(content string) tea.View {
	view := tea.NewView(look.Paint(content))
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

// Padded makes every line of a drawn screen exactly w cells wide: a design draws
// only what it has, and a program that prints the frame expects full lines.
func Padded(view string, w int) string {
	lines := strings.Split(view, "\n")
	for i, l := range lines {
		lines[i] = cells.PadRight(l, w)
	}
	return strings.Join(lines, "\n")
}
