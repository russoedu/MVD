package download

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// marqueeGap is the blank cells between the end of a scrolling text and its start again.
const marqueeGap = 4

// marqueeEvery is how many ticks the text waits before it moves one cell, and marqueePause is
// how many moves the start of the text is kept still, so it can be read before it goes.
const (
	marqueeEvery = 3
	marqueePause = 6
)

// marquee is text shown in avail cells: all of it when it fits, else a window that moves
// along it, and wraps round, as the ticks go by. The window is plain text: it is for the
// selected row, which the screen draws in one style.
func marquee(text string, avail, tick int) string {
	if avail <= 0 || ansi.StringWidth(text) <= avail {
		return text
	}
	loop := text + spaces(marqueeGap) + text
	cycle := ansi.StringWidth(text) + marqueeGap
	step := max(tick/marqueeEvery-marqueePause, 0) % (cycle + marqueePause)
	if step > cycle {
		step = 0
	}
	return ansi.Cut(loop, step, step+avail)
}

func spaces(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}

// wrapText breaks text into lines of at most w cells, at the spaces.
func wrapText(text string, w int) []string {
	return strings.Split(ansi.Wrap(text, max(w, 4), ""), string(rune(10)))
}
