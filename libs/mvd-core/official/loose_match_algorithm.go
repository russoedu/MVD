package official

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
	song := normalize(bracketed.ReplaceAllString(title, " "))
	if song == "" {
		return ""
	}
	original := normalize(title)

	for _, res := range results {
		if res.ID == "" {
			continue
		}
		resTitle := normalize(res.Title)
		if !containsWords(resTitle, song) || !fromAnyArtist(res, resTitle, artists) {
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
	channel := normalize(res.Channel)
	for _, artist := range artists {
		band := normalize(artist)
		if band != "" && (containsWords(resTitle, band) || containsWords(channel, band)) {
			return true
		}
	}
	return false
}

func hasLooseUnwanted(resultTitle, originalTitle string) bool {
	for _, word := range looseUnwanted {
		if containsWords(resultTitle, word) && !containsWords(originalTitle, word) {
			return true
		}
	}
	return false
}
