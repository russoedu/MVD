package official

import (
	"regexp"
	"strings"
)

// FoundTrack is the YouTube video chosen for a track.
type FoundTrack struct {
	VideoID string
	// Official is true when the video comes from the artist: their music video,
	// or an official audio or lyric upload of their own channel.
	Official bool
}

// FindTrack looks a song up on YouTube by its title and artists (as a
// playlist from another service lists them, "A, B" or "A & B" for several) and
// returns the best video: the official one when there is one, else the best
// non-official upload. durationSec is the length of the song, or 0 when
// unknown. It returns an empty VideoID when nothing matches.
func FindTrack(search Searcher, title, artists string, durationSec int, logf func(format string, a ...interface{})) (FoundTrack, error) {
	names := artistNames(artists)
	song := Song{Title: title, Artists: names, DurationSec: durationSec}

	pick, ok, results, err := FindBestVideo(search, song, logf)
	if err != nil {
		return FoundTrack{}, err
	}
	if ok {
		return FoundTrack{VideoID: pick.ID, Official: true}, nil
	}

	// Nothing official turned up: take the best other upload of the song.
	if id := PickLooseResult(results, title, names); id != "" {
		logf("only a non-official upload: %s", id)
		return FoundTrack{VideoID: id}, nil
	}

	logf("nothing that matches %q", strings.TrimSpace(title+" "+artists))
	return FoundTrack{}, nil
}

// artistJoiners join the names of artists in one credit.
var artistJoiners = regexp.MustCompile(`(?i)\s+(?:&|and|feat\.?|featuring|ft\.?|with|x|\+)\s+`)

// artistNames lists the artists a service names for a song, the first being
// the one to search for. "A, B" gives A and B. "A & B" gives "A & B" as
// written, since it may be one act ("Simon & Garfunkel"), then A and B, so a
// video that names only one of them still matches; the same goes for "and",
// "feat.", "with" and the like, so "Prince and the Revolution" also gives
// Prince, whose channel has the video.
func artistNames(artists string) []string {
	var names []string
	for _, name := range strings.Split(artists, ",") {
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}

	segments := len(names)
	for _, segment := range names[:segments] {
		parts := artistJoiners.Split(segment, -1)
		if len(parts) < 2 {
			continue
		}
		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				names = append(names, part)
			}
		}
	}
	return names
}
