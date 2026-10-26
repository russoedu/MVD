package catalog

import (
	"context"
	"testing"
)

func TestDeezerNamesArtistsAlbumsAndSongs(t *testing.T) {
	server := serving(t, map[string]string{
		"/search/artist":       `{"data":[{"id":27,"name":"Daft Punk"}]}`,
		"/artist/27/albums":    `{"data":[{"id":302127,"title":"Discovery","release_date":"2001-03-12"}]}`,
		"/album/302127/tracks": `{"data":[{"title":"One More Time","duration":320,"track_position":2,"disk_number":1,"artist":{"id":27,"name":"Daft Punk"}}]}`,
		"/search/track":        `{"data":[{"title":"Around the World","duration":429,"artist":{"id":27,"name":"Daft Punk"},"album":{"id":1,"title":"Homework"}}]}`,
		"/search/album":        `{"data":[{"id":302127,"title":"Discovery","artist":{"id":27,"name":"Daft Punk"}}]}`,
	})
	source := &Deezer{HTTP: server.Client(), BaseURL: server.URL}
	ctx := context.Background()

	artists, err := source.Artists(ctx, "Daft Punk")
	if err != nil || len(artists) != 1 || artists[0].ID != "27" {
		t.Fatalf("artists %+v, %v", artists, err)
	}
	albums, err := source.Albums(ctx, artists[0])
	if err != nil || len(albums) != 1 || albums[0].Artist != "Daft Punk" || albums[0].Released != "2001-03-12" {
		t.Fatalf("albums %+v, %v (an album of an artist is the artist's)", albums, err)
	}
	tracks, err := source.AlbumTracks(ctx, albums[0])
	if err != nil || len(tracks) != 1 || tracks[0].Album != "Discovery" || tracks[0].DurationMs != 320000 {
		t.Fatalf("tracks %+v, %v (a song of an album is on it)", tracks, err)
	}
	songs, err := source.SearchSongs(ctx, "Daft Punk", "Around the World")
	if err != nil || len(songs) != 1 || songs[0].Album != "Homework" {
		t.Fatalf("songs %+v, %v", songs, err)
	}
	found, err := source.SearchAlbums(ctx, "Daft Punk", "Discovery")
	if err != nil || len(found) != 1 || found[0].Title != "Discovery" {
		t.Fatalf("albums %+v, %v", found, err)
	}
}

func TestDeezerSaysItsProblemsInsideAnAnswer(t *testing.T) {
	server := serving(t, map[string]string{"/search/artist": `{"error":{"type":"Exception","message":"Quota limit exceeded","code":4}}`})
	if _, err := (&Deezer{HTTP: server.Client(), BaseURL: server.URL}).Artists(context.Background(), "x"); err == nil {
		t.Error("an error inside an answer was taken for no artists")
	}
}
