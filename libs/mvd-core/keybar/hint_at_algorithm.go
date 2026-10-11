package keybar

import "github.com/charmbracelet/x/ansi"

// At returns the key bar entry drawn under column x. The bar is a space, then each entry
// as its key, a space and its description, two spaces apart, exactly as the key bar of
// every design draws it (the test against the generated HintsHintAt of each screen keeps
// the two equal).
func At(hints []Hint, x int) (Hint, bool) {
	col := 1
	for i, h := range hints {
		if i > 0 {
			col += 2
		}
		width := ansi.StringWidth(h.Key)
		if h.Desc != "" {
			width += 1 + ansi.StringWidth(h.Desc)
		}
		if x >= col && x < col+width {
			return h, true
		}
		col += width
	}
	return Hint{}, false
}
