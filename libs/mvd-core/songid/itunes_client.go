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
}

// NewITunesClient returns a client pointed at the real iTunes Search API.
func NewITunesClient() *ITunesClient {
	return &ITunesClient{HTTP: &http.Client{Timeout: 20 * time.Second}, Endpoint: "https://itunes.apple.com/search"}
}

// Songs returns the songs the store finds for a free text, best match first.
func (c *ITunesClient) Songs(term string) ([]Identity, error) {
	resp, err := c.HTTP.Get(c.Endpoint + "?entity=song&limit=5&term=" + url.QueryEscape(term))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
