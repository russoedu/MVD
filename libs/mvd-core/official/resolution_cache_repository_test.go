package official

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheRemembersAcrossRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "official.json")
	song := Song{Title: "Uptown Girl (Remastered)", Artists: []string{"Billy Joel"}}

	first := NewResolutionCache(path, time.Hour)
	if _, ok := first.Get(song); ok {
		t.Fatal("an empty cache remembers nothing")
	}
	first.Put(song, Pick{ID: "hCuMWrfXG4E", Kind: KindVideo, Channel: "billyjoelVEVO", Why: "title 100%"})

	// A new cache over the same file: the next run.
	second := NewResolutionCache(path, time.Hour)
	got, ok := second.Get(Song{Title: "uptown girl", Artists: []string{"billy joel"}})
	if !ok || got.ID != "hCuMWrfXG4E" || got.Channel != "billyjoelVEVO" || got.Kind != KindVideo {
		t.Errorf("got %+v, %v; want the video remembered, whatever the spelling of the song", got, ok)
	}
}

func TestCacheForgetsWhatIsStaleOrIsTheSongsOwnUpload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "official.json")
	now := time.Now()
	cache := NewResolutionCache(path, 24*time.Hour)
	cache.now = func() time.Time { return now }

	song := Song{Title: "Song", Artists: []string{"Band"}}
	cache.Put(song, Pick{ID: "abcdefghijk"})

	cache.now = func() time.Time { return now.Add(23 * time.Hour) }
	if _, ok := cache.Get(song); !ok {
		t.Error("a day-old entry is fresh within its ttl")
	}
	cache.now = func() time.Time { return now.Add(25 * time.Hour) }
	if _, ok := cache.Get(song); ok {
		t.Error("an entry past its ttl is forgotten")
	}

	cache.now = func() time.Time { return now }
	own := song
	own.OwnID = "abcdefghijk"
	if _, ok := cache.Get(own); ok {
		t.Error("the song's own upload is never the answer")
	}
}

func TestCacheSurvivesABrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "official.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := NewResolutionCache(path, time.Hour)
	song := Song{Title: "Song", Artists: []string{"Band"}}
	if _, ok := cache.Get(song); ok {
		t.Error("a broken file is an empty cache")
	}
	cache.Put(song, Pick{ID: "abcdefghijk"})
	if _, ok := NewResolutionCache(path, time.Hour).Get(song); !ok {
		t.Error("the cache writes a good file over a broken one")
	}
}

func TestCacheKey(t *testing.T) {
	a := CacheKey(Song{Title: "Them Bones (2022 Remaster)", Artists: []string{"Alice In Chains"}})
	b := CacheKey(Song{Title: "them bones", Artists: []string{"ALICE IN CHAINS"}})
	if a != b {
		t.Errorf("%q and %q should be the same song", a, b)
	}
	if a == CacheKey(Song{Title: "Them Bones", Artists: []string{"Another Band"}}) {
		t.Error("the same title by another artist is another song")
	}
}
