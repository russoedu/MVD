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

func (c *ITunes) Name() string { return "iTunes" }

func (c *ITunes) ask(ctx context.Context, path string, query url.Values) (itunesAnswer, error) {
	var answer itunesAnswer
	err := getJSON(ctx, c.HTTP, c.BaseURL+path+"?"+query.Encode(), &answer)

	return answer, err
}

func (c *ITunes) Artists(ctx context.Context, name string) ([]Artist, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {name}, "entity": {"musicArtist"}, "limit": {"10"}})

	return answer.artists(), err
}

func (c *ITunes) Albums(ctx context.Context, artist Artist) ([]Album, error) {
	answer, err := c.ask(ctx, "/lookup", url.Values{"id": {artist.ID}, "entity": {"album"}, "limit": {"200"}})

	return answer.albums(), err
}

func (c *ITunes) AlbumTracks(ctx context.Context, album Album) ([]Track, error) {
	answer, err := c.ask(ctx, "/lookup", url.Values{"id": {album.ID}, "entity": {"song"}, "limit": {"200"}})

	return answer.tracks(), err
}

func (c *ITunes) SearchAlbums(ctx context.Context, artist, title string) ([]Album, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {artist + " " + title}, "entity": {"album"}, "limit": {"15"}})

	return answer.albums(), err
}

func (c *ITunes) SearchSongs(ctx context.Context, artist, title string) ([]Track, error) {
	answer, err := c.ask(ctx, "/search", url.Values{"term": {artist + " " + title}, "entity": {"song"}, "limit": {"25"}})

	return answer.tracks(), err
}
