package dialog

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Overlay lays popup over base with its top left corner at column x, row y. Both are
// screens of lines that may carry colours; the lines of base the popup does not reach, and
// the cells around it, are kept as they are, so the screen behind shows through at its
// sides.
func Overlay(base, popup string, x, y int) string {
	lines := strings.Split(base, "\n")
	pop := strings.Split(popup, "\n")
	x = max(x, 0)
	for i, line := range pop {
		row := y + i
		if row < 0 || row >= len(lines) {
			continue
		}
		w := ansi.StringWidth(line)
		under := lines[row]
		if have := ansi.StringWidth(under); have < x+w {
			under += strings.Repeat(" ", x+w-have)
		}
		lines[row] = ansi.Truncate(under, x, "") + "\x1b[0m" + line + "\x1b[0m" + ansi.TruncateLeft(under, x+w, "")
	}
	return strings.Join(lines, "\n")
}

// Centre is where a popup of w by h cells goes to be in the middle of a screen of
// screenW by screenH, never above or left of the corner.
func Centre(screenW, screenH, w, h int) (x, y int) {
	return max((screenW-w)/2, 0), max((screenH-h)/2, 0)
}

// Dim makes a drawn screen faint, to push it behind a popup laid over it: every line is set
// faint at its start, and again after each reset in it.
func Dim(screen string) string {
	const faint = "\x1b[2m"
	reset := strings.NewReplacer(
		"\x1b[0m", "\x1b[0m"+faint,
		"\x1b[m", "\x1b[m"+faint,
		"\x1b[22m", "\x1b[22m"+faint,
	)
	lines := strings.Split(screen, "\n")
	for i, l := range lines {
		lines[i] = faint + reset.Replace(l) + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}
