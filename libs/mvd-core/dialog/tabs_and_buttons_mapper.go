package dialog

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/look"
)

// Span is the columns a thing takes in a line, from X0 up to but not including X1.
type Span struct{ X0, X1 int }

// Contains reports whether column x is in the span.
func (s Span) Contains(x int) bool { return x >= s.X0 && x < s.X1 }

// Tabs draws a row of tabs with the active one lit, and says which columns each takes.
func Tabs(names []string, active int) (string, []Span) {
	var b strings.Builder
	spans := make([]Span, len(names))
	x := 0
	for i, name := range names {
		label := " " + name + " "
		spans[i] = Span{X0: x, X1: x + ansi.StringWidth(label)}
		if i == active {
			b.WriteString(look.Title.Reverse(true).Render(label))
		} else {
			b.WriteString(look.Dim.Render(label))
		}
		b.WriteString(" ")
		x += ansi.StringWidth(label) + 1
	}
	return b.String(), spans
}

// Buttons draws a row of buttons, the one under focus lit, and says which columns each
// takes. A button whose entry in disabled is true is greyed out.
func Buttons(labels []string, focus int, disabled map[string]bool) (string, []Span) {
	var b strings.Builder
	spans := make([]Span, len(labels))
	x := 0
	for i, name := range labels {
		label := "[ " + name + " ]"
		spans[i] = Span{X0: x, X1: x + ansi.StringWidth(label)}
		switch {
		case disabled[name]:
			b.WriteString(look.Dim.Faint(true).Render(label))
		case i == focus:
			b.WriteString(look.Title.Reverse(true).Render(label))
		default:
			b.WriteString(look.Yellow.Render(label))
		}
		b.WriteString("  ")
		x += ansi.StringWidth(label) + 2
	}
	return strings.TrimRight(b.String(), " "), spans
}

// At is the index of the span that holds column x, or -1.
func At(spans []Span, x int) int {
	for i, s := range spans {
		if s.Contains(x) {
			return i
		}
	}
	return -1
}
