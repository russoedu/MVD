package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/catalog"
	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/official/trackfind"
)

func TestSongFileProviderListsTheSongsOfAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Road trip.txt")
	if err := os.WriteFile(path, []byte("ATB - Killer\nSeal - Kiss From a Rose\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	source := newPlaylistTrackSource(nil, trackfind.Sources{})
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

// knownMusic is a catalogue that knows one artist with one song.
type knownMusic struct{ catalog.Source }

func (knownMusic) Name() string { return "known" }
func (knownMusic) Artists(context.Context, string) ([]catalog.Artist, error) {
	return []catalog.Artist{{ID: "1", Name: "ATB"}}, nil
}
func (knownMusic) Albums(context.Context, catalog.Artist) ([]catalog.Album, error) {
	return []catalog.Album{{ID: "a", Title: "Movin' Melodies", Artist: "ATB"}}, nil
}
func (knownMusic) AlbumTracks(context.Context, catalog.Album) ([]catalog.Track, error) {
	return []catalog.Track{{Artist: "ATB", Album: "Movin' Melodies", Title: "9 PM (Till I Come)", DurationMs: 215000}}, nil
}

func TestASearchAddressIsListedThroughTheCatalogueAndAUrlIsNot(t *testing.T) {
	source := newPlaylistTrackSource(nil, trackfind.Sources{})
	source.providers = append([]playlistProvider{searchProvider{finder: catalog.NewFinder(knownMusic{})}}, source.providers...)

	address := musicsearch.Query{Artist: "ATB"}.Address()
	if !source.Handles(address) || source.Handles("https://www.youtube.com/watch?v=x") || source.Handles("ATB - Killer") {
		t.Fatal("only the address of a search is the search provider's")
	}
	title, tracks, err := source.Tracks(context.Background(), address)
	if err != nil || title != "ATB" || len(tracks) != 1 || tracks[0].Title != "9 PM (Till I Come)" || tracks[0].Artist != "ATB" || tracks[0].DurationMs != 215000 {
		t.Fatalf("got %q, %+v, %v", title, tracks, err)
	}

	if _, _, err := source.Tracks(context.Background(), musicsearch.Scheme+"album=Lungs"); err == nil {
		t.Error("a search with no artist was listed")
	}
}
