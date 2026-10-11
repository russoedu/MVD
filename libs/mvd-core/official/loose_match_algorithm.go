package official

import (
	"youtube-downloader/libs/mvd-core/official/titlematch"
)

// looseUnwanted are the cuts that are never the song people mean, even when
// no official video exists. Lyric and audio uploads are fine here.
var looseUnwanted = []string{
	"cover", "karaoke", "live", "reaction", "instrumental", "remix", "slowed",
	"sped up", "8d", "nightcore", "reverb", "tutorial",
}

// PickLooseResult chooses the best result that is not the official video: it
// must carry the song title and one of the artists (by title or channel), and
// may be a lyric video, an audio upload or the artist's Topic channel, but
// not a cover, live, remix or other cut the song is not. It returns "" when
// none qualifies.
func PickLooseResult(results []SearchResult, title string, artists []string) string {
	song := titlematch.Normalize(titlematch.Bracketed.ReplaceAllString(title, " "))
	if song == "" {
		return ""
	}
	original := titlematch.Normalize(title)

	for _, res := range results {
		if res.ID == "" {
			continue
		}
		resTitle := titlematch.Normalize(res.Title)
		if !titlematch.ContainsWords(resTitle, song) || !fromAnyArtist(res, resTitle, artists) {
			continue
		}
		if hasLooseUnwanted(resTitle, original) {
			continue
		}
		return res.ID
	}
	return ""
}

// fromAnyArtist reports whether the result's title or channel names one of the
// artists. With no artists listed any result will do.
func fromAnyArtist(res SearchResult, resTitle string, artists []string) bool {
	if len(artists) == 0 {
		return true
	}
	channel := titlematch.Normalize(res.Channel)
	for _, artist := range artists {
		band := titlematch.Normalize(artist)
		if band != "" && (titlematch.ContainsWords(resTitle, band) || titlematch.ContainsWords(channel, band)) {
			return true
		}
	}
	return false
}

func hasLooseUnwanted(resultTitle, originalTitle string) bool {
	for _, word := range looseUnwanted {
		if titlematch.ContainsWords(resultTitle, word) && !titlematch.ContainsWords(originalTitle, word) {
			return true
		}
	}
	return false
}
