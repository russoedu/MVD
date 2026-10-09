package songid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// minMusicBrainzScore is the match score (0-100) below which an answer is ignored.
const minMusicBrainzScore = 85

// MusicBrainzClient searches the recordings of MusicBrainz, a free open database.
type MusicBrainzClient struct {
	HTTP      *http.Client
	Endpoint  string // the recording search endpoint
	UserAgent string // MusicBrainz requires one that says who is asking
	// MinInterval is the least time between two requests: MusicBrainz allows one a second.
	MinInterval time.Duration

	mu   sync.Mutex
	last time.Time
}

// NewMusicBrainzClient returns a client pointed at the real MusicBrainz.
func NewMusicBrainzClient() *MusicBrainzClient {
	return &MusicBrainzClient{
		HTTP:        &http.Client{Timeout: 20 * time.Second},
		Endpoint:    "https://musicbrainz.org/ws/2/recording",
		UserAgent:   "MVD-MusicVideoDownloader/1.0 (https://github.com/russoedu/MVD)",
		MinInterval: 1100 * time.Millisecond,
	}
}

// Recordings returns the recordings called title that artist made, best match first.
func (c *MusicBrainzClient) Recordings(artist, title string) ([]Identity, error) {
	query := fmt.Sprintf(`recording:"%s" AND artist:"%s"`, luceneEscaped(title), luceneEscaped(artist))
	c.wait()
	req, err := http.NewRequest(http.MethodGet, c.Endpoint+"?fmt=json&limit=5&query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MusicBrainz answered %s", resp.Status)
	}

	var body struct {
		Recordings []struct {
			Title        string `json:"title"`
			Score        int    `json:"score"`
			ArtistCredit []struct {
				Name string `json:"name"`
			} `json:"artist-credit"`
		} `json:"recordings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var out []Identity
	for _, rec := range body.Recordings {
		if rec.Score < minMusicBrainzScore || len(rec.ArtistCredit) == 0 {
			continue
		}
		out = append(out, Identity{Artist: rec.ArtistCredit[0].Name, Title: rec.Title})
	}
	return out, nil
}

// wait keeps the requests a polite distance apart.
func (c *MusicBrainzClient) wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if gap := c.MinInterval - time.Since(c.last); gap > 0 {
		time.Sleep(gap)
	}
	c.last = time.Now()
}

// luceneEscaped makes a text safe inside a quoted Lucene phrase.
func luceneEscaped(text string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(text)
}
