package downloadlist

import (
	"bytes"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

// encode is the list as the file would hold it, to tell whether it has changed. The
// creation time is left out: Save sets it the first time, and it is not a change.
func encode(f playlistfile.File) []byte {
	f.Created = f.Created.UTC().Truncate(0)
	var b bytes.Buffer
	if err := playlistfile.Write(&b, f); err != nil {
		return nil
	}
	return b.Bytes()
}

// copyOf is a list that shares nothing with f.
func copyOf(f playlistfile.File) playlistfile.File {
	out := playlistfile.File{Created: f.Created, Items: make([]playlistfile.Item, len(f.Items))}
	for i, item := range f.Items {
		out.Items[i] = playlistfile.Item{Source: item.Source, Songs: append([]playlistfile.Entry(nil), item.Songs...)}
	}
	return out
}
