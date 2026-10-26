package catalog

import (
	"strings"

	"youtube-downloader/libs/mvd-core/official/titlematch"
)

const (
	// minNameScore is how like an artist's name a catalogue's spelling must be to be the artist.
	minNameScore = 0.85
	// minSongScore is how well a catalogue's title must name the song that was asked for.
	minSongScore = 0.8
	// minAlbumScore is how well a catalogue's title must name the album that was asked for.
	minAlbumScore = 0.7
)

// unwantedWords make a song or an album one that is not the music of the artist: a karaoke backing, an
// instrumental or a tribute that carries the artist's name.
var unwantedWords = []string{"karaoke", "instrumental", "instrumentals", "tribute", "backing track", "in the style of"}

// versionWords make a song or an album a version of the music rather than the music: it is taken only
// when nothing better is there.
var versionWords = []string{"live", "remix", "remixes", "remixed", "acoustic", "demo", "demos", "bootleg", "mix"}

// artistIsNamed says whether a catalogue's artist is the one that was asked for: the same name, or one
// that only a slip or the spacing separates ("AC/DC" and "ACDC", "Guns N' Roses" and "Guns N Roses").
func artistIsNamed(asked, found string) bool {
	if askedKey := titlematch.CompactName(asked); askedKey != "" && askedKey == titlematch.CompactName(found) {
		return true
	}

	return titlematch.TitleSimilarity(asked, found, nil) >= minNameScore
}

// artistIsCredited says whether the artists of a song or an album include the one that was asked for:
// "Daft Punk & Pharrell Williams" does, and "Daft Punk Tribute Band" does not count.
func artistIsCredited(asked, credit string) bool {
	if artistIsNamed(asked, credit) {
		return true
	}
	if titlematch.ContainsWords(titlematch.Normalize(credit), titlematch.Normalize(asked)) {
		return !isUnwanted(credit)
	}

	return false
}

// songScore says how well a title names the song that was asked for, from 0 to 1.
func songScore(asked, found, artist string) float64 {
	return titlematch.TitleSimilarity(asked, found, []string{artist})
}

// hasWord says whether a text has one of the words, on word boundaries.
func hasWord(text string, words []string) bool {
	normal := titlematch.Normalize(text)
	for _, word := range words {
		if titlematch.ContainsWords(normal, word) {
			return true
		}
	}

	return false
}

// isUnwanted says whether a text names music that is not the artist's own.
func isUnwanted(text string) bool { return hasWord(text, unwantedWords) }

// versionRank is 1 for a text that names a version of the music (a live recording, a remix), 0 for the
// music itself.
func versionRank(text string) int {
	if hasWord(text, versionWords) || strings.Contains(strings.ToLower(text), "greatest hits") {
		return 1
	}

	return 0
}
