package catalog

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ITunes is the iTunes Search API, which needs no key.
type ITunes struct {
	HTTP    *http.Client
	BaseURL string
}

// NewITunes is the client of iTunes.
func NewITunes() *ITunes {
	return &ITunes{HTTP: &http.Client{Timeout: 20 * time.Second}, BaseURL: "https://itunes.apple.com"}
}

// Name returns "iTunes", the name of this catalogue.
func (c *ITunes) Name() string { return "iTunes" }

func (c *ITunes) ask(ctx context.Context, path string, query url.Values) (itunesAnswer, error) {
	var answer itunesAnswer
	err := getJSON(ctx, c.HTTP, c.BaseURL+path+"?"+query.Encode(), &answer)

	return answer, err
}

// Artists searches iTunes for artists named name.
// It returns at most 10 artists, or an error when the request fails.
func (c *ITunes) Artists(ctx context.Context, name string) ([]Artist, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {name}, "entity": {"musicArtist"}, "limit": {"10"}})

	return answer.artists(), err
}

// Albums lists the albums of artist, up to 200.
// It returns an error when the request fails.
func (c *ITunes) Albums(ctx context.Context, artist Artist) ([]Album, error) {
	answer, err := c.ask(ctx, "/lookup", url.Values{"id": {artist.ID}, "entity": {"album"}, "limit": {"200"}})

	return answer.albums(), err
}

// AlbumTracks lists the songs of album, up to 200.
// It returns an error when the request fails.
func (c *ITunes) AlbumTracks(ctx context.Context, album Album) ([]Track, error) {
	answer, err := c.ask(ctx, "/lookup", url.Values{"id": {album.ID}, "entity": {"song"}, "limit": {"200"}})

	return answer.tracks(), err
}

// SearchAlbums searches iTunes for albums called title by artist, up to 15.
// It returns an error when the request fails.
func (c *ITunes) SearchAlbums(ctx context.Context, artist, title string) ([]Album, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {artist + " " + title}, "entity": {"album"}, "limit": {"15"}})

	return answer.albums(), err
}

// SearchSongs searches iTunes for songs called title by artist, up to 25.
// It returns an error when the request fails.
func (c *ITunes) SearchSongs(ctx context.Context, artist, title string) ([]Track, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {artist + " " + title}, "entity": {"song"}, "limit": {"25"}})

	return answer.tracks(), err
}
