package songid

import (
	"regexp"
	"strings"

	"youtube-downloader/libs/mvd-core/official"
)

var (
	bracketPart  = regexp.MustCompile(`\(([^)]*)\)|\[([^\]]*)\]`)
	dashSplitter = regexp.MustCompile(`\s+[-–—]\s+`)
)

// guessesFrom lists the readings of an upload's title and channel worth asking a
// database about, the likeliest first: the artist named in the brackets, the two
// halves of "A - B" each way round, and the channel as the artist.
func guessesFrom(uploadTitle, channel string) []Identity {
	var guesses []Identity
	seen := map[string]bool{}
	add := func(artist, title string) {
		artist, title = strings.TrimSpace(artist), official.SearchTitle(title)
		key := strings.ToLower(artist + "|" + title)
		if artist == "" || title == "" || seen[key] {
			return
		}
		seen[key] = true
		guesses = append(guesses, Identity{Artist: artist, Title: title})
	}

	outside := strings.TrimSpace(bracketPart.ReplaceAllString(uploadTitle, " "))
	outside = strings.Join(strings.Fields(outside), " ")

	for _, match := range bracketPart.FindAllStringSubmatch(uploadTitle, -1) {
		inside := match[1] + match[2]
		if parts := dashSplitter.Split(inside, 2); len(parts) == 2 {
			add(parts[0], outside)
		}
	}
	if parts := dashSplitter.Split(outside, 2); len(parts) == 2 {
		add(parts[0], parts[1])
		add(parts[1], parts[0])
	}
	add(official.ArtistFromChannel(channel), outside)

	return guesses
}
