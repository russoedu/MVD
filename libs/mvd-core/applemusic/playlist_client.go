package applemusic

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

// Client fetches public playlists from Apple Music.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	// Rewrite, when set, changes the address fetched for a playlist link. Tests
	// use it to point the client at a local server.
	Rewrite func(playlistURL string) string
}

// NewClient returns a client pointed at the real Apple Music.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
}

// Fetch reads the public playlist behind an Apple Music link.
func (c *Client) Fetch(ctx context.Context, link string) (Playlist, error) {
	address, ok := PlaylistURL(link)
	if !ok {
		return Playlist{}, errors.New(link + " is not an Apple Music playlist link")
	}
	if c.Rewrite != nil {
		address = c.Rewrite(address)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
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
		return Playlist{}, errors.New("Apple Music has no such playlist (is it private?)")
	}
	if resp.StatusCode != http.StatusOK {
		return Playlist{}, errors.New("Apple Music answered " + resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return Playlist{}, err
	}
	return ParsePlaylistPage(string(body))
}
