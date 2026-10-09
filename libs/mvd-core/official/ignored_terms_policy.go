package official

import "strings"

// ignoredTerms are the words a title adds to name a version or a credit of a song
// rather than the song itself ("Radio Edit", "Extended Mix", "12” Version"). A
// search leaves them out, because a video called just "Run To You" is the song
// whatever mix the playlist's upload was.
var ignoredTerms = map[string]bool{
	"mix": true, "mixes": true, "remix": true, "edit": true, "version": true, "rework": true,
	"extended": true, "radio": true, "club": true, "dance": true, "dub": true, "vocal": true,
	"short": true, "long": true, "full": true, "album": true, "single": true, "original": true,
	"us": true, "uk": true, "euro": true, "european": true, "reprise": true, "clip": true,
	"audio": true, "lyrics": true, "lyric": true, "remaster": true, "remastered": true,
}

// versionTerms are the ignored terms that on their own say a text is a version of
// the song ("Radio Edit"), which "Dance" or "Club" are not ("Last Dance").
var versionTerms = map[string]bool{"mix": true, "mixes": true, "remix": true, "edit": true, "version": true, "rework": true}

// isIgnoredWord reports whether a normalized word is one a search leaves out.
func isIgnoredWord(word string) bool {
	return ignoredTerms[word] || titleNoise[word] || yearToken.MatchString(word) || strings.Trim(word, "0123456789") == ""
}
