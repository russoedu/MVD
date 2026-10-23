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
	return frame(title, lines, w, h, false)
}

// ClosableFrame is Frame with a close mark in the top right corner, which takes the
// columns CloseSpan says.
func ClosableFrame(title string, lines []string, w, h int) string {
	return frame(title, lines, w, h, true)
}

// closeMark is what the close button of a frame looks like.
const closeMark = " × "

// CloseSpan is the columns, in a ClosableFrame w cells wide, that close it when clicked.
func CloseSpan(w int) Span { return Span{X0: max(w, 4) - 5, X1: max(w, 4) - 2} }

// Sized is the size of a popup that takes percent of a screen in each direction, but not
// less than minW by minH cells, and never more than the screen less a margin of one cell.
func Sized(screenW, screenH, percent, minW, minH int) (w, h int) {
	w = min(max(screenW*percent/100, minW), max(screenW-2, 1))
	h = min(max(screenH*percent/100, minH), max(screenH-2, 1))
	return w, h
}

func frame(title string, lines []string, w, h int, closable bool) string {
	w, h = max(w, 4), max(h, 3)
	border := look.BorderOn
	inner := w - 2
	tail := ""
	if closable {
		tail = closeMark + "─"
	}
	room := inner - 1 - ansi.StringWidth(tail)

	head := ""
	if title != "" {
		text := " " + title + " "
		if ansi.StringWidth(text) > room-1 {
			text = ansi.Truncate(text, max(room-1, 1), "…")
		}
		head = border.Render("╭") + border.Render("─") + look.Title.Render(text) + border.Render(strings.Repeat("─", max(room-ansi.StringWidth(text), 0)))
	} else {
		head = border.Render("╭") + border.Render(strings.Repeat("─", max(room+1, 0)))
	}
	if closable {
		head += look.Red.Render(closeMark) + border.Render("─")
	}
	head += border.Render("╮")

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
