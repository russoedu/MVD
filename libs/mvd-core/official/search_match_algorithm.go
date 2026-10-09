package official

import (
	"regexp"
	"strings"
	"unicode"
)

var bracketed = regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]`)

// unwantedVersions are cuts that are never the official music video unless
// the art track itself is that cut.
var unwantedVersions = []string{
	"lyric", "lyrics", "audio", "cover", "karaoke", "live", "reaction", "instrumental",
	"remix", "slowed", "sped up", "8d", "nightcore", "reverb", "tutorial",
}

// SearchQuery is the text to search for the official video of a track.
func SearchQuery(title, artist string) string {
	return strings.TrimSpace(title + " " + artist + " official video")
}

// ArtistFromChannel strips the " - Topic" suffix from an art track channel.
func ArtistFromChannel(channel string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(channel), " - Topic"))
}

// PickSearchResult chooses the result that is most likely the official
// music video of the track, or "" when none is convincing. A result must
// carry the song title, come from the artist (by channel or title), not be
// an auto-generated track and not be a lyric, live or remix cut the art
// track is not. Results titled "official" win over the rest, the artist's own
// channel winning among them; a result that does not say "official" is taken
// only when the artist's own channel uploaded it.
func PickSearchResult(results []SearchResult, title, artist, ownID string) string {
	id, _ := pickSearchResult(results, title, artist, ownID)
	return id
}

// pickSearchResult is PickSearchResult that also says whether the result it
// chose is titled "official".
func pickSearchResult(results []SearchResult, title, artist, ownID string) (string, bool) {
	song := normalize(bracketed.ReplaceAllString(title, " "))
	if song == "" {
		return "", false
	}
	band := normalize(artist)
	original := normalize(title)

	bestID, bestTier, bestOfficial := "", 0, false
	for _, res := range results {
		if res.ID == "" || res.ID == ownID || IsTopicChannel(res.Channel) {
			continue
		}
		resTitle := normalize(res.Title)
		if !containsWords(resTitle, song) {
			continue
		}
		if band != "" && !containsWords(resTitle, band) && !containsWords(normalize(res.Channel), band) && !ChannelIsArtist(res.Channel, artist) {
			continue
		}
		if hasUnwantedVersion(resTitle, original) {
			continue
		}
		// A video that says "official" from the artist's own channel is the best pick;
		// one that says "official" from anyone else's channel comes next (fans write
		// it too); one that does not say it counts only when the artist uploaded it.
		official := containsWords(resTitle, "official")
		artistChannel := ChannelIsArtist(res.Channel, artist)
		tier := 0
		switch {
		case official && artistChannel:
			tier = 3
		case official:
			tier = 2
		case artistChannel:
			tier = 1
		}
		if tier > bestTier {
			bestID, bestTier, bestOfficial = res.ID, tier, official
		}
	}
	return bestID, bestOfficial
}

func hasUnwantedVersion(resultTitle, originalTitle string) bool {
	for _, word := range unwantedVersions {
		if containsWords(resultTitle, word) && !containsWords(originalTitle, word) {
			return true
		}
	}
	return false
}

// normalize lowercases and reduces text to single-space separated words.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
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
