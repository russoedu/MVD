package official

import (
	"regexp"
	"strings"
	"unicode"
)

var bracketed = regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]`)

// ArtistFromChannel strips the " - Topic" suffix from an art track channel.
func ArtistFromChannel(channel string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(channel), " - Topic"))
}

// normalize lowercases and reduces text to single-space separated words,
// without accents.
func normalize(s string) string {
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

// containsWords reports whether needle appears in haystack on word boundaries.
func containsWords(haystack, needle string) bool {
	return strings.Contains(" "+haystack+" ", " "+needle+" ")
}
