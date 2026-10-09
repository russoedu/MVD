package songfile

import (
	"regexp"
	"strings"
)

// listNumber is the "12." or "3)" some lists put before each song.
var listNumber = regexp.MustCompile(`^\d+[.)]\s+`)

// dividers separate the artist from the title on a line.
var dividers = []string{" - ", " – ", " — ", "\t"}

// ParseLines reads "Artist - Title" lines. Blank lines and # comments are
// ignored; a line without a divider is counted in the second result.
func ParseLines(text string) ([]Song, int) {
	var songs []Song
	skipped := 0
	for _, line := range strings.Split(strings.TrimPrefix(text, "\ufeff"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = listNumber.ReplaceAllString(line, "")

		artist, title, ok := splitLine(line)
		if !ok {
			skipped++
			continue
		}
		songs = append(songs, Song{Title: title, Artist: artist})
	}
	return songs, skipped
}

// splitLine cuts a line at its first divider.
func splitLine(line string) (artist, title string, ok bool) {
	best := -1
	length := 0
	for _, divider := range dividers {
		if i := strings.Index(line, divider); i >= 0 && (best < 0 || i < best) {
			best, length = i, len(divider)
		}
	}
	if best < 0 {
		return "", "", false
	}
	artist = strings.TrimSpace(line[:best])
	title = strings.TrimSpace(line[best+length:])
	return artist, title, artist != "" && title != ""
}
