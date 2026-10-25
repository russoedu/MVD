package songlist

import (
	"strings"

	"youtube-downloader/libs/mvd-core/runstate"
)

// VideoTitle is the title of the video of a song, or its id in brackets when the listing had none.
func VideoTitle(en *runstate.Entry) string {
	if t := strings.TrimSpace(en.Title); t != "" {
		return t
	}

	return "(" + en.VideoID + ")"
}

// Name is the song as it is called: as a music database named it, else the title of its video.
func Name(en *runstate.Entry) string {
	if en.NamedTitle != "" {
		return en.NamedTitle
	}

	return VideoTitle(en)
}

// Artist is who sings a song, or the channel it came from when no artist is known.
func Artist(en *runstate.Entry) string {
	switch {
	case en.NamedArtist != "":
		return en.NamedArtist
	case en.Channel != "":
		// A channel of the music service is named after the artist, with this added.
		return strings.TrimSuffix(en.Channel, " - Topic")
	}

	return ""
}
