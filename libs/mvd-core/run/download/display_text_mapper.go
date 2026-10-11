package download

import (
	"youtube-downloader/libs/mvd-core/cells"

	"youtube-downloader/libs/mvd-core/runstate"
)

// entryTitle returns the title to show for an entry, falling back to the
// video id when the listing had no title.
func entryTitle(en *runstate.Entry) string {
	t := cells.Display(en.Title)
	if t == "" {
		return "(" + en.VideoID + ")"
	}
	return t
}
