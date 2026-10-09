package official

import (
	"errors"
	"strings"
	"testing"
)

// scriptedSearch answers each query with the results listed for it (matched by
// a substring of the query) and remembers what was asked.
type scriptedSearch struct {
	answers map[string][]SearchResult
	asked   []string
}

func (s *scriptedSearch) search(query string) ([]SearchResult, error) {
	s.asked = append(s.asked, query)
	for part, results := range s.answers {
		if strings.Contains(query, part) {
			return results, nil
		}
	}
	return nil, nil
}

func TestSearchQueries(t *testing.T) {
	song := Song{Title: "Them Bones (2022 Remaster)", Artists: []string{"Alice In Chains"}}
	got := searchQueries(song)
	want := []string{
		"Them Bones Alice In Chains official video",
		"Alice In Chains Them Bones official music video",
		"Alice In Chains - Them Bones",
	}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("query %d = %q, want %q", i, got[i], want[i])
		}
	}

	if queries := searchQueries(Song{Title: "Song"}); len(queries) == 0 || strings.Contains(queries[0], "  ") {
		t.Errorf("without an artist the queries are still clean: %q", queries)
	}
}

func TestFindBestVideoStopsAtAConfidentMusicVideo(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official video": {{ID: "official", Title: "Seal - Killer (Official Video)", Channel: "Seal", Verified: true, Views: 1400000}},
	}}
	song := Song{Title: "Killer", Artists: []string{"Seal"}}

	pick, ok, _, err := FindBestVideo(search.search, Sources{}, song, quietLog)
	if err != nil || !ok || pick.ID != "official" {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if len(search.asked) != 1 {
		t.Errorf("a confident video needs one search, asked %q", search.asked)
	}
}

func TestFindBestVideoTriesOtherWordingsWhenTheFirstFindsNothing(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official music video": {{ID: "official", Title: "Hanson - MMMBop (Official Music Video)", Channel: "HANSON", Verified: true}},
	}}
	song := Song{Title: "MMMBop", Artists: []string{"Hanson"}}

	pick, ok, _, err := FindBestVideo(search.search, Sources{}, song, quietLog)
	if err != nil || !ok || pick.ID != "official" {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if len(search.asked) != 2 {
		t.Errorf("asked %q, want the second wording to be tried", search.asked)
	}
}

func TestFindBestVideoKeepsTheBestAcrossSearchesAndReportsTheResults(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"Lovely Day Bill Withers official video": {{ID: "fan", Title: "Bill Withers - Lovely Day", Channel: "Memory Lane"}},
		"Bill Withers - Lovely Day":              {{ID: "audio", Title: "Lovely Day (Official Audio)", Channel: "Bill Withers"}},
	}}
	song := Song{Title: "Lovely Day", Artists: []string{"Bill Withers"}}

	pick, ok, results, err := FindBestVideo(search.search, Sources{}, song, quietLog)
	if err != nil || !ok || pick.ID != "audio" || pick.Kind != KindAudio {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if len(results) != 2 || len(search.asked) != 3 {
		t.Errorf("want 2 results from 3 searches, got %d from %q", len(results), search.asked)
	}
}

func TestFindBestVideoErrors(t *testing.T) {
	failing := func(string) ([]SearchResult, error) { return nil, errors.New("offline") }
	if _, _, _, err := FindBestVideo(failing, Sources{}, Song{Title: "Song", Artists: []string{"Band"}}, quietLog); err == nil {
		t.Error("the first search failing is an error")
	}
	if _, _, _, err := FindBestVideo(nil, Sources{}, Song{Title: "Song"}, quietLog); err == nil {
		t.Error("no searcher is an error")
	}

	// A later search failing only ends the search.
	calls := 0
	flaky := func(string) ([]SearchResult, error) {
		calls++
		if calls == 1 {
			return []SearchResult{{ID: "fan", Title: "Band - Song", Channel: "Fan"}}, nil
		}
		return nil, errors.New("offline")
	}
	if _, ok, _, err := FindBestVideo(flaky, Sources{}, Song{Title: "Song", Artists: []string{"Band"}}, quietLog); err != nil || ok {
		t.Errorf("got ok=%v err=%v, want no pick and no error", ok, err)
	}
}
