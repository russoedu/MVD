package dialog

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/look"
)

// Frame draws a box w columns wide with a rounded border and the title in its top edge.
// Each of lines is cut or padded to fit inside it, and the box is h rows tall, the
// border included: lines the box has no room for are dropped, and missing ones are blank.
func Frame(title string, lines []string, w, h int) string {
	w, h = max(w, 4), max(h, 3)
	border := look.BorderOn
	inner := w - 2

	head := ""
	if title != "" {
		text := " " + title + " "
		if ansi.StringWidth(text) > inner-2 {
			text = ansi.Truncate(text, inner-2, "…")
		}
		head = border.Render("╭") + border.Render("─") + look.Title.Render(text) + border.Render(strings.Repeat("─", max(inner-1-ansi.StringWidth(text), 0))) + border.Render("╮")
	} else {
		head = border.Render("╭" + strings.Repeat("─", inner) + "╮")
	}

	out := make([]string, 0, h)
	out = append(out, head)
	for i := 0; i < h-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		out = append(out, border.Render("│")+fit(line, inner)+border.Render("│"))
	}
	out = append(out, border.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(out, "\n")
}

// fit cuts or pads a line to exactly w cells.
func fit(s string, w int) string {
	if n := ansi.StringWidth(s); n > w {
		return ansi.Truncate(s, w, "")
	} else {
		return s + strings.Repeat(" ", w-n)
	}
}

// Fit is fit, for the pages that fill a Frame.
func Fit(s string, w int) string { return fit(s, w) }
