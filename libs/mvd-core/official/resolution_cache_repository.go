package official

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CachedPick is a video remembered for a song.
type CachedPick struct {
	ID      string    `json:"id"`
	Kind    Kind      `json:"kind"`
	Channel string    `json:"channel"`
	Why     string    `json:"why"`
	At      time.Time `json:"at"`
}

// ResolutionCache remembers the video found for a song, in a JSON file, so a
// second run over the same playlist asks nobody. Only videos are remembered:
// a song without one is searched for again, since one may appear.
type ResolutionCache struct {
	path string
	ttl  time.Duration
	now  func() time.Time

	mu      sync.Mutex
	entries map[string]CachedPick
	loaded  bool
}

// NewResolutionCache returns a cache stored at path whose entries are good for ttl.
func NewResolutionCache(path string, ttl time.Duration) *ResolutionCache {
	return &ResolutionCache{path: path, ttl: ttl, now: time.Now, entries: map[string]CachedPick{}}
}

// CacheKey is what identifies a song: its title and artists, in a spelling that
// does not depend on case, accents or punctuation.
func CacheKey(song Song) string {
	return normalize(SearchTitle(song.Title)) + "|" + normalize(strings.Join(song.Artists, ","))
}

// Get returns the video remembered for a song, if it is still fresh.
func (c *ResolutionCache) Get(song Song) (CachedPick, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.load()

	pick, ok := c.entries[CacheKey(song)]
	if !ok || c.now().Sub(pick.At) > c.ttl || pick.ID == song.OwnID {
		return CachedPick{}, false
	}
	return pick, true
}

// Put remembers the video found for a song and saves the file. A file that
// cannot be written only means the next run asks again.
func (c *ResolutionCache) Put(song Song, pick Pick) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.load()

	c.entries[CacheKey(song)] = CachedPick{ID: pick.ID, Kind: pick.Kind, Channel: pick.Channel, Why: pick.Why, At: c.now()}
	c.save()
}

// load reads the file once; a file that is missing or unreadable is an empty cache.
func (c *ResolutionCache) load() {
	if c.loaded {
		return
	}
	c.loaded = true

	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	var stored map[string]CachedPick
	if json.Unmarshal(data, &stored) == nil {
		c.entries = stored
	}
}

// save writes the file by way of a temporary one, so a crash never leaves half a file.
func (c *ResolutionCache) save() {
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return
	}
	temporary := c.path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return
	}
	if err := os.Rename(temporary, c.path); err != nil {
		_ = os.Remove(temporary)
	}
}
