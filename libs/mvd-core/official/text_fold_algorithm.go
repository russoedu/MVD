package official

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// foldAccents takes the accents off letters, so "ADÉLA" and "Adela", "Beyoncé"
// and "Beyonce" are the same name.
func foldAccents(s string) string {
	folder := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	folded, _, err := transform.String(folder, s)
	if err != nil {
		return s
	}
	return folded
}
