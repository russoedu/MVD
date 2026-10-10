package playlisteditor

import (
	"strings"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Columns are the parts of a song the editor shows side by side.
type Columns struct {
	// Artist and Song are what a music database named, or else what the upload's own
	// title and channel say.
	Artist, Song string
	// Video is the title of the upload in the playlist.
	Video string
	// Replaced is the title of the video the person chose instead; empty without one.
	Replaced string
}

// ColumnsOf splits a song into the editor's columns. A song a database named keeps its
// names. Otherwise "Artist - Song" in the title is split, and a title that carries no
// artist takes the channel's name, without YouTube's " - Topic".
func ColumnsOf(r playlistfile.Entry) Columns {
	c := Columns{Artist: r.Artist, Song: r.Title, Video: r.Upload.Title}
	if c.Artist == "" || c.Song == "" {
		artist, song := splitTitle(r.Upload.Title)
		if artist == "" {
			artist = strings.TrimSuffix(strings.TrimSpace(r.Upload.Channel), " - Topic")
		}
		if c.Artist == "" {
			c.Artist = artist
		}
		if c.Song == "" {
			c.Song = song
		}
	}
	if c.Song == "" {
		c.Song = "(" + r.Upload.ID + ")"
	}
	if r.Decision == playlistfile.DecisionReplaced {
		c.Replaced = r.ChosenTitle
		if c.Replaced == "" {
			c.Replaced = "(" + r.ChosenID + ")"
		}
	}
	return c
}

// splitTitle splits "Artist - Song" at its first " - "; a title without one is all song.
func splitTitle(title string) (artist, song string) {
	title = strings.TrimSpace(title)
	if left, right, found := strings.Cut(title, " - "); found && left != "" && right != "" {
		return strings.TrimSpace(left), strings.TrimSpace(right)
	}
	return "", title
}
