package catalog

import "context"

// Artist is an artist of a catalogue.
type Artist struct {
	ID   string
	Name string
}

// Album is an album of a catalogue.
type Album struct {
	ID     string
	Artist string
	Title  string
	// Released is the date it came out as the catalogue gives it (it sorts as text), or "".
	Released string
}

// Track is a song of a catalogue.
type Track struct {
	// Artist is the artists as the catalogue credits them.
	Artist     string
	Album      string
	Title      string
	DurationMs int
	// Disc and Number place the song on its album; 0 when unknown.
	Disc, Number int
}

// Source is a place that knows music. Each call asks it once; a source that cannot answer says so with
// an error and the finder goes on to the next.
type Source interface {
	// Name is what the source is called, for the words of an error.
	Name() string
	// Artists are the artists that go by name, best first.
	Artists(ctx context.Context, name string) ([]Artist, error)
	// Albums are the albums of an artist.
	Albums(ctx context.Context, artist Artist) ([]Album, error)
	// AlbumTracks are the songs of an album, in the order of the album.
	AlbumTracks(ctx context.Context, album Album) ([]Track, error)
	// SearchAlbums are the albums that may be the one the artist and the title name, best first.
	SearchAlbums(ctx context.Context, artist, title string) ([]Album, error)
	// SearchSongs are the songs that may be the one the artist and the title name, best first.
	SearchSongs(ctx context.Context, artist, title string) ([]Track, error)
}
