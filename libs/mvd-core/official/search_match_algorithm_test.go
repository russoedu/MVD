package official

import "testing"

func TestPickSearchResult(t *testing.T) {
	results := []SearchResult{
		{ID: "topic", Title: "Wonderwall", Channel: "Oasis - Topic"},
		{ID: "lyrics", Title: "Oasis - Wonderwall (Lyrics)", Channel: "LyricsCo"},
		{ID: "cover", Title: "Wonderwall cover by Someone", Channel: "Someone"},
		{ID: "other", Title: "Wonderwall", Channel: "Random Guy"},
		{ID: "fan", Title: "Oasis - Wonderwall", Channel: "Fan Upload"},
		{ID: "official", Title: "Oasis - Wonderwall (Official Video)", Channel: "Oasis"},
	}
	if got := PickSearchResult(results, "Wonderwall", "Oasis", "own"); got != "official" {
		t.Errorf("got %q, want the official upload", got)
	}
	if got := PickSearchResult(results[:5], "Wonderwall", "Oasis", "own"); got != "fan" {
		t.Errorf("got %q, want the best non-official match", got)
	}
	if got := PickSearchResult(results[:4], "Wonderwall", "Oasis", "own"); got != "" {
		t.Errorf("got %q, want no match without artist evidence", got)
	}
	if got := PickSearchResult(results, "Wonderwall", "Oasis", "official"); got == "official" {
		t.Error("must not pick the track's own video")
	}
}

func TestPickSearchResultKeepsWantedVersions(t *testing.T) {
	results := []SearchResult{{ID: "live", Title: "Oasis - Wonderwall (Live)", Channel: "Oasis"}}
	if got := PickSearchResult(results, "Wonderwall (Live)", "Oasis", "own"); got != "live" {
		t.Errorf("a live track should accept a live video, got %q", got)
	}
	if got := PickSearchResult(results, "Wonderwall", "Oasis", "own"); got != "" {
		t.Errorf("a studio track must reject a live video, got %q", got)
	}
}

func TestSearchQuery(t *testing.T) {
	if got := SearchQuery("Wonderwall", "Oasis"); got != "Wonderwall Oasis official video" {
		t.Errorf("got %q", got)
	}
	if got := ArtistFromChannel("Oasis - Topic"); got != "Oasis" {
		t.Errorf("got %q", got)
	}
}
