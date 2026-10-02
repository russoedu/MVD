package tui

import (
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
	"golang.org/x/text/unicode/norm"

	"youtube-downloader/internal/runstate"
)

// display makes a user supplied string safe to draw in a fixed layout.
//
// Terminals do not agree on how many cells an emoji takes (Windows ConPTY,
// xterm.js and the width tables used here can all differ), and a single
// disagreement makes a line wrap and shifts every later repaint by one
// row. So pictographs are dropped, tabs become spaces, control and format
// characters are removed, and the text is normalised so accented letters
// are single code points. East Asian text is kept: its width is stable.
func display(s string) string {
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

// entryTitle returns the title to show for an entry, falling back to the
// video id when the listing had no title.
func entryTitle(en *runstate.Entry) string {
	t := display(en.Title)
	if t == "" {
		return "(" + en.VideoID + ")"
	}
	return t
}
