package official

import (
	"regexp"
	"strings"
)

// creditSuffix is the credit of a featured artist after the title ("Love Is Gone feat. Jane").
var creditSuffix = regexp.MustCompile(`(?i)\s+(?:feat|ft|featuring)\.?(?:\s.*)?$`)

// SearchTitle is a song's title as it is searched for: without the bracketed
// parts, the credit of a featured artist, and the version words (see
// ignoredTerms) after a dash or at the end, which only narrow a search down to
// nothing ("Them Bones (2022 Remaster)" is searched as "Them Bones").
func SearchTitle(title string) string {
	cleaned := strings.TrimSpace(bracketed.ReplaceAllString(title, " "))
	cleaned = creditSuffix.ReplaceAllString(cleaned, "")
	if head, tail, found := strings.Cut(cleaned, " - "); found && hasNoise(tail) {
		cleaned = head
	}
	cleaned = withoutTrailingVersion(strings.Join(strings.Fields(cleaned), " "))
	if cleaned == "" {
		return strings.TrimSpace(title)
	}
	return cleaned
}

// withoutTrailingVersion cuts the version words that end a title ("Pump It Up Radio
// Edit" is "Pump It Up"), as long as a word that names a version is among them and
// something of the title is left.
func withoutTrailingVersion(title string) string {
	words := strings.Fields(title)
	keep, named := len(words), false
	for keep > 1 && isIgnoredWord(normalize(words[keep-1])) {
		if versionTerms[normalize(words[keep-1])] {
			named = true
		}
		keep--
	}
	if !named || keep == len(words) {
		return title
	}
	return strings.Join(words[:keep], " ")
}

// hasNoise reports whether a text is made of version words and years only.
func hasNoise(text string) bool {
	words := strings.Fields(normalize(text))
	if len(words) == 0 {
		return false
	}
	for _, word := range words {
		if !isIgnoredWord(word) {
			return false
		}
	}
	return true
}
