// Package cells measures and shapes text in terminal cells: cutting, wrapping and padding
// a line to a width, and making a user supplied string safe to draw in a fixed layout.
package cells

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"golang.org/x/text/unicode/norm"
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

// Display makes a user supplied string safe to draw in a fixed layout.
//
// Terminals do not agree on how many cells an emoji takes (Windows ConPTY,
// xterm.js and the width tables used here can all differ), and a single
// disagreement makes a line wrap and shifts every later repaint by one
// row. So pictographs are dropped, tabs become spaces, control and format
// characters are removed, and the text is normalised so accented letters
// are single code points. East Asian text is kept: its width is stable.
func Display(s string) string {
	if s == "" {
		return s
	}
	s = norm.NFC.String(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteByte(' ')
		case r == '\n' || r == '\r':
			b.WriteByte(' ')
		case unicode.Is(unicode.Cc, r), unicode.Is(unicode.Cf, r), unicode.Is(unicode.Co, r), unicode.Is(unicode.Cs, r):
			// control, format (ZWJ, bidi marks), private use, surrogates
		case r >= 0xFE00 && r <= 0xFE0F, r == 0x20E3:
			// variation selectors and the keycap combiner
		case r >= 0x1F000:
			// emoji and pictographs
		case r >= 0x2600 && r <= 0x27BF && runewidth.RuneWidth(r) == 2:
			// miscellaneous symbols and dingbats that render as emoji
		case runewidth.RuneWidth(r) == 2 && !eastAsian(r):
			// any other double width symbol
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// eastAsian reports whether r belongs to a script whose double width
// rendering is consistent across terminals.
func eastAsian(r rune) bool {
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Bopomofo, r) {
		return true
	}
	// CJK symbols and punctuation, fullwidth forms
	return (r >= 0x3000 && r <= 0x303F) || (r >= 0xFF00 && r <= 0xFFEF)
}
