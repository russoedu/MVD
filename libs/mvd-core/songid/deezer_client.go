package songid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DeezerClient searches the tracks of Deezer's public catalogue, which needs no account.
type DeezerClient struct {
	HTTP     *http.Client
	Endpoint string // the search endpoint
	// MinInterval is the least time between two requests, however many workers ask.
	MinInterval time.Duration

	pace pacer
}

// NewDeezerClient returns a client pointed at the real Deezer API.
func NewDeezerClient() *DeezerClient {
	return &DeezerClient{HTTP: &http.Client{Timeout: 20 * time.Second}, Endpoint: "https://api.deezer.com/search", MinInterval: 120 * time.Millisecond}
}

// Tracks returns the tracks Deezer finds for a free text, best match first.
func (c *DeezerClient) Tracks(term string) ([]Identity, error) {
	if c.pace.blocked() {
		return nil, errLeftAlone
	}
	c.pace.wait(c.MinInterval)
	resp, err := c.HTTP.Get(c.Endpoint + "?limit=5&q=" + url.QueryEscape(term))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		if refusedBy(resp.StatusCode) {
			c.pace.block(leftAloneFor)
		}
		return nil, fmt.Errorf("the Deezer search answered %s", resp.Status)
	}
	var body struct {
		Data []struct {
			Title  string `json:"title"`
			Artist struct {
				Name string `json:"name"`
			} `json:"artist"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var out []Identity
	for _, t := range body.Data {
		if t.Title != "" && t.Artist.Name != "" {
			out = append(out, Identity{Artist: t.Artist.Name, Title: t.Title})
		}
	}
	return out, nil
}
