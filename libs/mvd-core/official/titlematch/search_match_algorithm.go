package titlematch

import (
	"regexp"
	"strings"
	"unicode"
)

// Bracketed matches a part of a title in parentheses or square brackets, such as (Official Video).
var Bracketed = regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]`)

// ArtistFromChannel strips the " - Topic" suffix from an art track channel.
func ArtistFromChannel(channel string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(channel), " - Topic"))
}

// Normalize lowercases and reduces text to single-space separated words,
// without accents.
func Normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(foldAccents(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// ContainsWords reports whether needle appears in haystack on word boundaries.
func ContainsWords(haystack, needle string) bool {
	return strings.Contains(" "+haystack+" ", " "+needle+" ")
}
