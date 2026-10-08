package official

import (
	"fmt"
	"strings"
)

// FoundTrack is the YouTube video chosen for a track.
type FoundTrack struct {
	VideoID string
	// Official is true when the video is titled as the official one.
	Official bool
}

// FindTrack looks a song up on YouTube by its title and artists (as a
// playlist from another service lists them, "A, B" for several) and returns
// the best video: the official one when there is one, else the best
// non-official upload. It returns an empty VideoID when nothing matches.
func FindTrack(search Searcher, title, artists string, logf func(format string, a ...interface{})) (FoundTrack, error) {
	if search == nil {
		return FoundTrack{}, fmt.Errorf("no way to search YouTube")
	}
	names := artistNames(artists)
	primary := ""
	if len(names) > 0 {
		primary = names[0]
	}

	// 1. The official video: "<title> <artist> official video".
	query := SearchQuery(title, primary)
	results, err := search(query)
	if err != nil {
		return FoundTrack{}, fmt.Errorf("search for %q failed: %w", query, err)
	}
	for _, artist := range names {
		if id, official := pickSearchResult(results, title, artist, ""); id != "" {
			logf("search for %q found %s", query, id)
			return FoundTrack{VideoID: id, Official: official}, nil
		}
	}
	if id := PickLooseResult(results, title, names); id != "" {
		logf("search for %q found only a non-official upload: %s", query, id)
		return FoundTrack{VideoID: id}, nil
	}

	// 2. No official video turned up: search for the song itself.
	plain := strings.TrimSpace(title + " " + primary)
	results, err = search(plain)
	if err != nil {
		return FoundTrack{}, fmt.Errorf("search for %q failed: %w", plain, err)
	}
	for _, artist := range names {
		if id, official := pickSearchResult(results, title, artist, ""); id != "" {
			logf("search for %q found %s", plain, id)
			return FoundTrack{VideoID: id, Official: official}, nil
		}
	}
	if id := PickLooseResult(results, title, names); id != "" {
		logf("search for %q found only a non-official upload: %s", plain, id)
		return FoundTrack{VideoID: id}, nil
	}

	logf("search for %q found nothing that matches", plain)
	return FoundTrack{}, nil
}

// artistNames splits "A, B" into its artists.
func artistNames(artists string) []string {
	var names []string
	for _, name := range strings.Split(artists, ",") {
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}
	return names
}
