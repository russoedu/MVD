package official

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// WikidataClient asks Wikidata which YouTube videos it lists for a song. Many
// songs have an item with the "YouTube video ID" property (P1651), usually the
// official video. It is an optional source: callers treat an error as "no answer".
type WikidataClient struct {
	HTTP      *http.Client
	Endpoint  string // the SPARQL endpoint
	UserAgent string // Wikidata asks for one that says who is asking
	// MinInterval is the least time between two requests, to be a polite client.
	MinInterval time.Duration

	mu   sync.Mutex
	last time.Time
}

// NewWikidataClient returns a client pointed at the real Wikidata.
func NewWikidataClient() *WikidataClient {
	return &WikidataClient{
		HTTP:        &http.Client{Timeout: 20 * time.Second},
		Endpoint:    "https://query.wikidata.org/sparql",
		UserAgent:   "MVD-MusicVideoDownloader/1.0 (https://github.com/russoedu/MVD)",
		MinInterval: time.Second,
	}
}

// VideoIDs returns the YouTube video ids Wikidata lists for songs titled like
// the song and performed by one of the artists. A song with the right title
// by someone else (a cover) is left out.
func (c *WikidataClient) VideoIDs(title string, artists []string) ([]string, error) {
	labels := titleLabels(title)
	if len(labels) == 0 || len(artists) == 0 {
		return nil, nil
	}

	query := `SELECT ?vid ?perfLabel WHERE {
  VALUES ?label { ` + strings.Join(labels, " ") + ` }
  ?item rdfs:label ?label .
  ?item wdt:P1651 ?vid .
  ?item wdt:P175 ?perf .
  ?perf rdfs:label ?perfLabel . FILTER(LANG(?perfLabel) = "en")
} LIMIT 50`

	c.wait()
	req, err := http.NewRequest(http.MethodGet, c.Endpoint+"?format=json&query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Wikidata answered " + resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	var answer struct {
		Results struct {
			Bindings []struct {
				Vid       struct{ Value string } `json:"vid"`
				PerfLabel struct{ Value string } `json:"perfLabel"`
			} `json:"bindings"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &answer); err != nil {
		return nil, err
	}

	var ids []string
	seen := map[string]bool{}
	for _, row := range answer.Results.Bindings {
		id := row.Vid.Value
		if !videoIDShape.MatchString(id) || seen[id] || !performedBy(row.PerfLabel.Value, artists) {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

// wait spaces the requests of one client.
func (c *WikidataClient) wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if gap := c.MinInterval - time.Since(c.last); gap > 0 {
		time.Sleep(gap)
	}
	c.last = time.Now()
}

// titleLabels are the SPARQL literals a song's label may be: the title as
// written and without its bracketed and version parts.
func titleLabels(title string) []string {
	var labels []string
	seen := map[string]bool{}
	for _, candidate := range []string{strings.TrimSpace(title), SearchTitle(title)} {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", " ").Replace(candidate)
		labels = append(labels, `"`+escaped+`"@en`)
	}
	return labels
}

// performedBy reports whether a performer's name is one of the artists.
func performedBy(performer string, artists []string) bool {
	key := strings.TrimPrefix(compactName(performer), "the")
	for _, artist := range artists {
		if other := strings.TrimPrefix(compactName(artist), "the"); key != "" && key == other {
			return true
		}
	}
	return false
}
