package official

import (
	"strings"
)

// SearchTitle is a song's title as it is searched for: without the bracketed
// parts and the version suffix after a dash, which only narrow a search down to
// nothing ("Them Bones (2022 Remaster)" is searched as "Them Bones").
func SearchTitle(title string) string {
	cleaned := strings.TrimSpace(bracketed.ReplaceAllString(title, " "))
	if head, tail, found := strings.Cut(cleaned, " - "); found && hasNoise(tail) {
		cleaned = head
	}
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	if cleaned == "" {
		return strings.TrimSpace(title)
	}
	return cleaned
}

// hasNoise reports whether a text is made of version words and years only.
func hasNoise(text string) bool {
	words := strings.Fields(normalize(text))
	if len(words) == 0 {
		return false
	}
	for _, word := range words {
		if !titleNoise[word] && !yearToken.MatchString(word) && strings.Trim(word, "0123456789") != "" {
			return false
		}
	}
	return true
}
