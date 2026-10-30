package catalog

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// Deezer is the public API of Deezer, which needs no key.
type Deezer struct {
	HTTP    *http.Client
	BaseURL string
}

// NewDeezer is the client of Deezer.
func NewDeezer() *Deezer {
	return &Deezer{HTTP: &http.Client{Timeout: 20 * time.Second}, BaseURL: "https://api.deezer.com"}
}

// Name returns "Deezer", the name of this catalogue.
func (c *Deezer) Name() string { return "Deezer" }

// deezerGet asks Deezer for a list; Deezer answers a problem with an error inside a 200.
func deezerGet[T any](ctx context.Context, c *Deezer, path string, query url.Values) ([]T, error) {
	var list deezerList[T]
	if err := getJSON(ctx, c.HTTP, c.BaseURL+path+"?"+query.Encode(), &list); err != nil {
		return nil, err
	}
	if msg := list.problem(); msg != "" {
		return nil, errors.New(msg)
	}

	return list.Data, nil
}

// Artists searches Deezer for artists named name.
// It returns at most 10 artists, or an error when the request fails or Deezer reports a problem.
func (c *Deezer) Artists(ctx context.Context, name string) ([]Artist, error) {
	rows, err := deezerGet[deezerArtist](ctx, c, "/search/artist", url.Values{"q": {name}, "limit": {"10"}})

	return artistsOf(rows), err
}

// Albums lists the albums of artist, up to 100.
// It returns an error when the request fails or Deezer reports a problem.
func (c *Deezer) Albums(ctx context.Context, artist Artist) ([]Album, error) {
	rows, err := deezerGet[deezerAlbum](ctx, c, "/artist/"+url.PathEscape(artist.ID)+"/albums", url.Values{"limit": {"100"}})

	return albumsOf(rows, artist.Name), err
}

// AlbumTracks lists the tracks of album, up to 200.
// A track with no artist takes the artist of the album.
// It returns an error when the request fails or Deezer reports a problem.
func (c *Deezer) AlbumTracks(ctx context.Context, album Album) ([]Track, error) {
	rows, err := deezerGet[deezerTrack](ctx, c, "/album/"+url.PathEscape(album.ID)+"/tracks", url.Values{"limit": {"200"}})
	out := tracksOf(rows, album.Title)
	for i := range out {
		if out[i].Artist == "" {
			out[i].Artist = album.Artist
		}
	}

	return out, err
}

// SearchAlbums searches Deezer for albums called title by artist, up to 15.
// It returns an error when the request fails or Deezer reports a problem.
func (c *Deezer) SearchAlbums(ctx context.Context, artist, title string) ([]Album, error) {
	rows, err := deezerGet[deezerAlbum](ctx, c, "/search/album", url.Values{"q": {`artist:"` + artist + `" album:"` + title + `"`}, "limit": {"15"}})

	return albumsOf(rows, ""), err
}

// SearchSongs searches Deezer for songs called title by artist, up to 25.
// It returns an error when the request fails or Deezer reports a problem.
func (c *Deezer) SearchSongs(ctx context.Context, artist, title string) ([]Track, error) {
	rows, err := deezerGet[deezerTrack](ctx, c, "/search/track", url.Values{"q": {`artist:"` + artist + `" track:"` + title + `"`}, "limit": {"25"}})

	return tracksOf(rows, ""), err
}
