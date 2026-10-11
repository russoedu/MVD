package help

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"
)

// Screen draws the help overlay for a window of width by height cells: the keys, the
// glyphs and where the full output is written, fitted to the room the window has. It
// is the hand written side of the Help page; the page itself is generated.
func Screen(width, height int, logPath string) string {
	maxW := width - 6
	if maxW < 20 {
		maxW = 20
	}
	var lines []string
	for _, l := range []string{
		look.Key.Render("↑ ↓ j k") + "     move selection, scroll output",
		look.Key.Render("pgup pgdn") + "   move ten rows",
		look.Key.Render("tab") + "         cycle panes: playlists, entries, output",
		look.Key.Render("enter") + "       toggle follow mode (jump to the active download)",
		look.Key.Render("f") + "           show only failed entries",
		look.Key.Render("r") + "           retry the selected failed entry, or all failed in the playlist",
		look.Key.Render("l") + "           toggle full screen output",
		look.Key.Render("a") + "           add more links to this run",
		look.Key.Render("q") + "           quit (asks for confirmation while running)",
		look.Key.Render("ctrl+c") + "      quit immediately",
		look.Key.Render("ctrl+l") + "      redraw the screen",
		"",
		look.Dim.Render("glyphs: · queued  ⟲ resolving  ⠋ downloading  ⚙ merging  ✓ done  ≡ duplicate  ✗ failed  ⇄ official video"),
		"",
		look.Dim.Render("full output is written to " + logPath),
		"",
		look.Dim.Render("press any key to close"),
	} {
		if ansi.StringWidth(l) <= maxW {
			lines = append(lines, l)
			continue
		}
		lines = append(lines, cells.Wrap(l, maxW)...)
	}
	if len(lines) > height-2 {
		lines = lines[:height-2]
	}
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	widest := 0
	for _, l := range lines {
		widest = max(widest, ansi.StringWidth(l))
	}
	props.HelpWidth, props.HelpHeight = widest+4, len(lines)+2
	props.Lines = nil
	for _, l := range lines {
		props.Lines = append(props.Lines, LinesRow{Line: cells.CellText(l)})
	}
	return padded(View(props, width, height).View, width)
}

// padded makes every line of a drawn page exactly w cells wide: a design draws only what
// it has, and a program that prints the frame expects full lines.
func padded(view string, w int) string {
	lines := strings.Split(view, "\n")
	for i, l := range lines {
		lines[i] = cells.PadRight(l, w)
	}
	return strings.Join(lines, "\n")
}
