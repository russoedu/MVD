package workspace

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// render draws the menu bar on the top row and, under it, the review or the download
// screen. An open menu is laid over them, and the question or the notice takes the bottom
// row.
// Render is the screen as text, for a host that lays something over it.
func (m Model) Render() string { return m.render() }

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	below := m.screen.Render()
	if m.reviewing {
		below = m.editor.Render()
	}
	lines := strings.Split(below, "\n")
	if len(lines) > m.height-topRows {
		lines = lines[:m.height-topRows]
	}
	if bottom := m.bottomLine(); bottom != "" && len(lines) > 0 {
		lines[len(lines)-1] = bottom
	}
	screen := m.barRow() + "\n" + m.toolbarView() + "\n" + strings.Join(lines, "\n")
	if m.editing {
		x, y := m.editOrigin()
		w, h := m.editSize()
		screen = dialog.Overlay(screen, m.edit.sized(w, h).View(), x, y)
	}
	if m.about {
		x, y := m.aboutOrigin()
		screen = dialog.Overlay(screen, m.aboutView(), x, y)
	}
	if m.bar.Open() >= 0 {
		screen = overlayMenu(screen, m.bar.View())
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

// overlayMenu lays the open menu over the screen: the bar on the top row, and the box of
// the menu alone under it, so the screen shows at both sides of the box and below it.
func overlayMenu(screen, barView string) string {
	lines := strings.Split(screen, "\n")
	rows := strings.Split(barView, "\n")
	lines[0] = rows[0]
	screen = strings.Join(lines, "\n")
	for y := 1; y < len(rows); y++ {
		plain := ansi.Strip(rows[y])
		left := len(plain) - len(strings.TrimLeft(plain, " "))
		right := ansi.StringWidth(strings.TrimRight(plain, " "))
		if right <= left {
			continue
		}
		screen = dialog.Overlay(screen, ansi.Cut(rows[y], left, right), left, y)
	}
	return screen
}
