package workspace

import (
	"strings"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// render draws the menu bar on the top row and, under it, the review or the download
// screen. An open menu is laid over them, and the question or the notice takes the bottom
// row.
func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	below := m.screen.Render()
	if m.reviewing {
		below = m.editor.Render()
	}
	lines := strings.Split(below, "\n")
	if len(lines) > m.height-1 {
		lines = lines[:m.height-1]
	}
	if bottom := m.bottomLine(); bottom != "" && len(lines) > 0 {
		lines[len(lines)-1] = bottom
	}
	screen := m.barRow() + "\n" + strings.Join(lines, "\n")
	if m.editing {
		w, h := min(m.width-4, 100), min(m.height-4, 24)
		x, y := dialog.Centre(m.width, m.height, w, h)
		screen = dialog.Overlay(screen, m.edit.View(w, h), x, y)
	}
	if m.bar.Open() >= 0 {
		screen = dialog.Overlay(screen, m.bar.View(), 0, 0)
	}
	return screen
}

// barRow is the top row: the bar alone.
func (m Model) barRow() string {
	view := m.bar
	view.Close()
	view.SetSize(m.width, 1)
	return view.View()
}

// bottomLine is the question being asked, else what the last action said.
func (m Model) bottomLine() string {
	switch {
	case m.asking != askNone:
		return cells.Fit(" "+look.Key.Render(m.input.View()), m.width)
	case m.starting:
		return cells.Fit(" "+look.Dim.Render("Getting ready..."), m.width)
	case m.stopping:
		return cells.Fit(" "+look.Dim.Render("Stopping the downloads..."), m.width)
	case m.notice != "":
		return cells.Fit(" "+look.Red.Render(m.notice), m.width)
	}
	return ""
}
