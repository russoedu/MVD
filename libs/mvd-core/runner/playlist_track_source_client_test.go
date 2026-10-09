package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/official"
)

func TestSongFileProviderListsTheSongsOfAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Road trip.txt")
	if err := os.WriteFile(path, []byte("ATB - Killer\nSeal - Kiss From a Rose\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	source := newPlaylistTrackSource(nil, official.Sources{})
	if !source.Handles(path) || source.Handles("https://www.youtube.com/playlist?list=PLx") {
		t.Fatalf("a song file is ours, a YouTube playlist is not")
	}
	title, tracks, err := source.Tracks(context.Background(), `"`+path+`"`)
	if err != nil || title != "Road trip" || len(tracks) != 2 || tracks[0].Artist != "ATB" || tracks[1].Title != "Kiss From a Rose" {
		t.Fatalf("got %q, %+v, %v", title, tracks, err)
	}

	if _, _, err := source.Tracks(context.Background(), filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Error("a missing file is an error")
	}
}
