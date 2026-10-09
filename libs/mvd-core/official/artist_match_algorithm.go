package official

import (
	"strings"
	"unicode"
)

// channelNoise are the words channels add around an artist's name: the Vevo
// channel of "Adele" is "AdeleVEVO", INXS posts as "INXS Videos".
var channelNoise = []string{"vevo", "official", "music", "records", "tv", "videos", "channel"}

// ChannelIsArtist reports whether a YouTube channel is the artist's own, by
// name: "Nickelback", "billyjoelVEVO" and "INXS Videos" are, "Trance Mission"
// is not the channel of ATB. "&" and "and" are the same, so "Earth, Wind & Fire"
// is the artist of "EarthWindandFireVEVO".
func ChannelIsArtist(channel, artist string) bool {
	// "Michael Sembello (The Master)" is Michael Sembello's.
	channelKey := stripNoise(compactName(bracketed.ReplaceAllString(channel, " ")))
	if channelKey == "" {
		return false
	}
	for _, key := range artistKeys(artist) {
		if key != "" && (channelKey == key || strings.TrimPrefix(channelKey, "the") == key) {
			return true
		}
	}
	return false
}

// artistKeys are the spellings an artist's name may have in a channel name.
func artistKeys(artist string) []string {
	withAnd := compactName(strings.ReplaceAll(artist, "&", " and "))
	withoutAnd := compactName(strings.ReplaceAll(artist, "&", " "))
	return []string{withAnd, withoutAnd, strings.TrimPrefix(withAnd, "the"), strings.TrimPrefix(withoutAnd, "the")}
}

// compactName lowercases a name and keeps only its letters and digits.
func compactName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, foldAccents(name))
}

// stripNoise takes the channel noise words off both ends of a compact name.
func stripNoise(key string) string {
	for changed := true; changed; {
		changed = false
		for _, noise := range channelNoise {
			if len(key) > len(noise)+2 && strings.HasSuffix(key, noise) {
				key, changed = key[:len(key)-len(noise)], true
			}
			if len(key) > len(noise)+2 && strings.HasPrefix(key, noise) {
				key, changed = key[len(noise):], true
			}
		}
	}
	return key
}
