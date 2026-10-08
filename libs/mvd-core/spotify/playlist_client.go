package spotify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client fetches public playlists from Spotify.
type Client struct {
	HTTP      *http.Client
	EmbedBase string // e.g. https://open.spotify.com/embed/playlist/
	UserAgent string
}

// NewClient returns a client pointed at the real Spotify.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		EmbedBase: "https://open.spotify.com/embed/playlist/",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
}

// Fetch reads the public playlist behind a Spotify link.
func (c *Client) Fetch(ctx context.Context, link string) (Playlist, error) {
	id, ok := PlaylistID(link)
	if !ok {
		return Playlist{}, fmt.Errorf("%q is not a Spotify playlist link", link)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.EmbedBase+id, nil)
	if err != nil {
		return Playlist{}, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Playlist{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return Playlist{}, errors.New("Spotify has no playlist " + id + " (is it private?)")
	}
	if resp.StatusCode != http.StatusOK {
		return Playlist{}, errors.New("Spotify answered " + resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return Playlist{}, err
	}
	return ParseEmbedPage(string(body))
}
