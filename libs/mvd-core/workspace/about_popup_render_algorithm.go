package workspace

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// aboutLayout is the About window drawn once: its lines, and where on them the link and the
// button are, counted from the left of the inside of the frame.
type aboutLayout struct {
	lines          []string
	linkY, buttonY int
	link, button   dialog.Span
	width, height  int
}

// aboutLayout draws the About window for a screen m.width by m.height.
func (m Model) aboutLayout() aboutLayout {
	w := min(max(m.width*70/100, 50), 84)
	w = min(w, max(m.width-2, 4))
	iw := w - 2
	text := max(iw-4, 10)

	version := m.host.Version
	if version == "" {
		version = "dev"
	}
	var lines []string
	add := func(s string) { lines = append(lines, "  "+s) }
	add("")
	add(look.Title.Render("MVD") + " " + look.Text.Render("· Music Video Downloader"))
	add(look.Dim.Render("Version ") + look.Yellow.Render(version))
	linkY := len(lines)
	add(look.Cyan.Underline(true).Render(projectURL))
	add("")
	for _, p := range aboutParagraphs {
		for _, row := range strings.Split(ansi.Wrap(p, text, ""), "\n") {
			add(look.Text.Render(row))
		}
		add("")
	}
	add(look.Dim.Italic(true).Render(aboutFooter))
	add("")
	buttonY := len(lines)
	_, spans := dialog.Buttons([]string{"OK"}, 0, nil)
	pad := max((iw-spans[0].X1)/2, 0)
	focus := -1
	if h := min(len(lines)+2, max(m.height-2, 4)); m.hoverY >= 0 {
		x0, y0 := dialog.Centre(m.width, m.height, w, h)
		if m.hoverY == y0+1+min(len(lines), h-3) && m.hoverX >= x0+1+pad && m.hoverX < x0+1+pad+spans[0].X1 {
			focus = 0
		}
	}
	label, _ := dialog.Buttons([]string{"OK"}, focus, nil)
	lines = append(lines, strings.Repeat(" ", pad)+label)

	h := min(len(lines)+2, max(m.height-2, 4))
	return aboutLayout{
		lines: lines, linkY: linkY, buttonY: buttonY, width: w, height: h,
		link:   dialog.Span{X0: 2, X1: 2 + ansi.StringWidth(projectURL)},
		button: dialog.Span{X0: pad + spans[0].X0, X1: pad + spans[0].X1},
	}
}

// aboutView is the About window.
func (m Model) aboutView() string {
	l := m.aboutLayout()
	lines := make([]string, 0, l.height-2)
	for i := 0; i < l.height-2 && i < len(l.lines); i++ {
		lines = append(lines, cells.Fit(l.lines[i], l.width-2))
	}
	// When the screen is too short for all of it, the button still shows.
	if len(l.lines) > l.height-2 && len(lines) > 0 {
		lines[len(lines)-1] = l.lines[l.buttonY]
	}
	return dialog.ClosableFrame("About MVD", lines, l.width, l.height)
}
