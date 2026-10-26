package catalog

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"youtube-downloader/libs/mvd-core/musicsearch"
)

// fakeSource is a catalogue with a few artists, albums and songs; a nil field says nothing, and err
// makes every call fail.
type fakeSource struct {
	name    string
	artists []Artist
	albums  map[string][]Album // by artist id
	tracks  map[string][]Track // by album id
	songs   []Track
	err     error
	calls   atomic.Int32
}

func (f *fakeSource) Name() string { return f.name }
func (f *fakeSource) Artists(context.Context, string) ([]Artist, error) {
	f.calls.Add(1)

	return f.artists, f.err
}
func (f *fakeSource) Albums(_ context.Context, a Artist) ([]Album, error) {
	return f.albums[a.ID], f.err
}
func (f *fakeSource) AlbumTracks(_ context.Context, a Album) ([]Track, error) {
	return f.tracks[a.ID], f.err
}
func (f *fakeSource) SearchAlbums(_ context.Context, _, _ string) ([]Album, error) {
	var all []Album
	for _, list := range f.albums {
		all = append(all, list...)
	}

	return all, f.err
}
func (f *fakeSource) SearchSongs(context.Context, string, string) ([]Track, error) {
	f.calls.Add(1)

	return f.songs, f.err
}

func daftPunk() *fakeSource {
	return &fakeSource{
		name:    "fake",
		artists: []Artist{{ID: "x", Name: "Daft Punk Tribute"}, {ID: "1", Name: "Daft Punk"}},
		albums: map[string][]Album{"1": {
			{ID: "live", Title: "Alive 2007 (Live)", Artist: "Daft Punk", Released: "2007-11-19"},
			{ID: "d", Title: "Discovery", Artist: "Daft Punk", Released: "2001-03-13"},
			{ID: "h", Title: "Homework", Artist: "Daft Punk", Released: "1997-01-20"},
			{ID: "k", Title: "Discovery (Karaoke Version)", Artist: "Daft Punk", Released: "2002-01-01"},
		}},
		tracks: map[string][]Track{
			"h":    {{Artist: "Daft Punk", Album: "Homework", Title: "Around the World", Number: 2}, {Artist: "Daft Punk", Album: "Homework", Title: "Da Funk", Number: 1}},
			"d":    {{Artist: "Daft Punk", Album: "Discovery", Title: "One More Time", Number: 1}, {Artist: "Daft Punk", Album: "Discovery", Title: "Around the World (Remastered)", Number: 2}},
			"live": {{Artist: "Daft Punk", Album: "Alive 2007", Title: "One More Time (Live)", Number: 1}, {Artist: "Daft Punk", Album: "Alive 2007", Title: "Robot Rock", Number: 2}},
		},
		songs: []Track{
			{Artist: "Daft Punk", Album: "Alive 2007 (Live)", Title: "One More Time (Live)", DurationMs: 1},
			{Artist: "Daft Punk Tribute Band", Album: "Covers", Title: "One More Time", DurationMs: 2},
			{Artist: "Daft Punk", Album: "Discovery", Title: "One More Time", DurationMs: 320357},
		},
	}
}

func titles(tracks []Track) string {
	out := make([]string, len(tracks))
	for i, t := range tracks {
		out[i] = t.Title
	}

	return strings.Join(out, " | ")
}

func TestAnArtistGivesEverySongOnceTheMusicBeforeItsVersions(t *testing.T) {
	res, err := NewFinder(daftPunk()).Find(context.Background(), musicsearch.Query{Artist: "daft punk"})
	if err != nil {
		t.Fatal(err)
	}
	want := "Da Funk | Around the World | One More Time | Robot Rock"
	if got := titles(res.Tracks); got != want {
		t.Errorf("songs %q, want %q", got, want)
	}
	if res.Title != "daft punk" {
		t.Errorf("title %q", res.Title)
	}
}

func TestAnAlbumGivesItsSongsInOrder(t *testing.T) {
	res, err := NewFinder(daftPunk()).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk", Album: "homework"})
	if err != nil || titles(res.Tracks) != "Da Funk | Around the World" {
		t.Fatalf("got %q, %v", titles(res.Tracks), err)
	}
	if _, err := NewFinder(daftPunk()).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk", Album: "Random Access Memories"}); err == nil || !strings.Contains(err.Error(), "no songs found") {
		t.Errorf("an album nobody knows: %v", err)
	}
}

func TestASongIsTheStudioSongOfTheArtistNotATributeOrALiveVersion(t *testing.T) {
	res, err := NewFinder(daftPunk()).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk", Song: "one more time"})
	if err != nil || len(res.Tracks) != 1 || res.Tracks[0].DurationMs != 320357 || res.Tracks[0].Album != "Discovery" {
		t.Fatalf("got %+v, %v", res.Tracks, err)
	}
	withAlbum, err := NewFinder(daftPunk()).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk", Album: "Alive 2007", Song: "One More Time"})
	if err != nil || len(withAlbum.Tracks) != 1 || withAlbum.Tracks[0].Album != "Alive 2007 (Live)" {
		t.Fatalf("with the live album named, its recording is the one: got %+v, %v", withAlbum.Tracks, err)
	}
}

func TestASongNoCatalogueKnowsIsTheSongThatWasTyped(t *testing.T) {
	res, err := NewFinder(&fakeSource{name: "empty"}).Find(context.Background(), musicsearch.Query{Artist: "Unknown Band", Song: "Garage Demo"})
	if err != nil || len(res.Tracks) != 1 || res.Tracks[0].Artist != "Unknown Band" || res.Tracks[0].Title != "Garage Demo" {
		t.Errorf("got %+v, %v", res.Tracks, err)
	}
}

func TestTheNextCatalogueIsAskedWhenTheFirstFailsOrDoesNotKnow(t *testing.T) {
	broken := &fakeSource{name: "broken", err: errors.New("down")}
	empty := &fakeSource{name: "empty"}
	good := daftPunk()
	res, err := NewFinder(broken, empty, good).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk"})
	if err != nil || len(res.Tracks) == 0 {
		t.Fatalf("got %+v, %v", res.Tracks, err)
	}
	if broken.calls.Load() == 0 || empty.calls.Load() == 0 {
		t.Error("the earlier catalogues were not asked first")
	}

	_, err = NewFinder(broken, &fakeSource{name: "also broken", err: errors.New("down too")}).Find(context.Background(), musicsearch.Query{Artist: "Daft Punk"})
	if err == nil || !strings.Contains(err.Error(), "cannot look up") {
		t.Errorf("every catalogue down: %v", err)
	}
}

func TestAQueryThatCannotBeSearchedIsRefused(t *testing.T) {
	good := daftPunk()
	if _, err := NewFinder(good).Find(context.Background(), musicsearch.Query{Song: "One More Time"}); err == nil || good.calls.Load() != 0 {
		t.Errorf("a song with no artist was searched: %v", err)
	}
}

func TestAnEndedContextEndsTheSearch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewFinder(&fakeSource{name: "x", err: ctx.Err()}).Find(ctx, musicsearch.Query{Artist: "Daft Punk"}); !errors.Is(err, context.Canceled) {
		t.Errorf("got %v", err)
	}
}
