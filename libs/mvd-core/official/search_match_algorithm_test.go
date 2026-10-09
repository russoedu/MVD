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
	if got := PickSearchResult(results[:5], "Wonderwall", "Oasis", "own"); got != "" {
		t.Errorf("got %q, a fan upload that does not say official is not taken", got)
	}
	if got := PickSearchResult(results[:4], "Wonderwall", "Oasis", "own"); got != "" {
		t.Errorf("got %q, want no match without artist evidence", got)
	}
	if got := PickSearchResult(results, "Wonderwall", "Oasis", "official"); got == "official" {
		t.Error("must not pick the track's own video")
	}
}

func TestPickSearchResultPrefersTheArtistsOwnChannel(t *testing.T) {
	// A video from the artist's channel that does not say "official" is taken
	// (Nickelback's own "How You Remind Me" has no such word).
	own := SearchResult{ID: "own-channel", Title: "Nickelback - How You Remind Me", Channel: "Nickelback"}
	fan := SearchResult{ID: "fan", Title: "Nickelback - How You Remind Me", Channel: "Fan Upload"}
	if got := PickSearchResult([]SearchResult{fan, own}, "How You Remind Me", "Nickelback", "x"); got != "own-channel" {
		t.Errorf("got %q, want the video of the artist's own channel", got)
	}

	// Fans write "official" too: among the ones that say it, the artist's channel wins
	// even when it ranks lower in the search.
	fanOfficial := SearchResult{ID: "fan-official", Title: "Real McCoy - Run Away (Official HD Video 1993)", Channel: "MusicBoxChannel"}
	ownOfficial := SearchResult{ID: "own-official", Title: "Real McCoy - Run Away (Official Video)", Channel: "Real McCoy"}
	if got := PickSearchResult([]SearchResult{fanOfficial, ownOfficial}, "Run Away", "Real McCoy", "x"); got != "own-official" {
		t.Errorf("got %q, want the official video of the artist's channel", got)
	}
	if got := PickSearchResult([]SearchResult{fanOfficial}, "Run Away", "Real McCoy", "x"); got != "fan-official" {
		t.Errorf("got %q, a fan's \"official\" upload is still taken when nothing better exists", got)
	}

	// A Vevo channel is the artist's.
	vevo := SearchResult{ID: "vevo", Title: "Uptown Girl", Channel: "billyjoelVEVO"}
	if got := PickSearchResult([]SearchResult{vevo}, "Uptown Girl", "Billy Joel", "x"); got != "vevo" {
		t.Errorf("got %q, want the Vevo channel's video although its title omits the artist", got)
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
