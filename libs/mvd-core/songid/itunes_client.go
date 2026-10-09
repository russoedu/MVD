package songid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ITunesClient searches the songs of the iTunes Store, which needs no account.
type ITunesClient struct {
	HTTP     *http.Client
	Endpoint string // the search endpoint
	// MinInterval is the least time between two requests, however many workers ask.
	MinInterval time.Duration

	pace pacer
}

// NewITunesClient returns a client pointed at the real iTunes Search API.
func NewITunesClient() *ITunesClient {
	return &ITunesClient{HTTP: &http.Client{Timeout: 20 * time.Second}, Endpoint: "https://itunes.apple.com/search", MinInterval: 300 * time.Millisecond}
}

// Songs returns the songs the store finds for a free text, best match first.
func (c *ITunesClient) Songs(term string) ([]Identity, error) {
	if c.pace.blocked() {
		return nil, errLeftAlone
	}
	c.pace.wait(c.MinInterval)
	resp, err := c.HTTP.Get(c.Endpoint + "?entity=song&limit=5&term=" + url.QueryEscape(term))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		if refusedBy(resp.StatusCode) {
			c.pace.block(leftAloneFor)
		}
		return nil, fmt.Errorf("iTunes answered %s", resp.Status)
	}
	var body struct {
		Results []struct {
			TrackName  string `json:"trackName"`
			ArtistName string `json:"artistName"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var out []Identity
	for _, r := range body.Results {
		if r.TrackName != "" && r.ArtistName != "" {
			out = append(out, Identity{Artist: r.ArtistName, Title: r.TrackName})
		}
	}
	return out, nil
}
