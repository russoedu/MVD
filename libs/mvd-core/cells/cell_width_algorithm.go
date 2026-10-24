// Package cells measures and shapes text in terminal cells: cutting, wrapping and padding
// a line to a width, and making a user supplied string safe to draw in a fixed layout.
package cells

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Truncate cuts s to at most w cells, ending in an ellipsis when it was cut.
func Truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= w {
		return s
	}
	if w == 1 {
		return ansi.Truncate(s, 1, "")
	}
	return ansi.Truncate(s, w, "…")
}

// Wrap breaks s into lines of at most w cells on word boundaries.
func Wrap(s string, w int) []string {
	if w <= 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for ansi.StringWidth(word) > w {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			lines = append(lines, ansi.Truncate(word, w, ""))
			word = word[len(ansi.Truncate(word, w, "")):]
		}
		switch {
		case line == "":
			line = word
		case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// TruncateMiddle cuts s to at most w cells by taking out the middle, so both ends stay: the start
// of a path says where it is, and its end says what it is. A third of the room goes to the start.
func TruncateMiddle(s string, w int) string {
	if w <= 0 {
		return ""
	}
	sw := ansi.StringWidth(s)
	if sw <= w {
		return s
	}
	if w <= 3 {
		return Truncate(s, w)
	}
	head := (w - 1) / 3
	tail := w - 1 - head
	return ansi.Truncate(s, head, "") + "…" + ansi.TruncateLeft(s, sw-tail, "")
}

// WrapLine breaks one line of text into rows of at most w cells, on word boundaries and in
// the middle of a word that is longer than a row. Unlike Wrap it keeps the spaces and the
// colours of the line, so a log line reads the same on two rows as on one.
func WrapLine(s string, w int) []string {
	if w <= 0 {
		return nil
	}
	if ansi.StringWidth(s) <= w {
		return []string{s}
	}
	return strings.Split(ansi.Wrap(s, w, ""), "\n")
}

// Fit pads or truncates s to exactly w cells.
func Fit(s string, w int) string {
	return PadRight(Truncate(s, w), w)
}

// PadRight pads s with spaces on the right up to w cells.
func PadRight(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

// CellText is the text of a row cell: a blank line is a space, since a row with nothing
// in it is not kept by a design's rows.
func CellText(s string) string {
	if s == "" {
		return " "
	}
	return s
}
