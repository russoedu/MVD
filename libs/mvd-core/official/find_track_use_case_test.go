package official

import (
	"errors"
	"testing"
)

func quiet(string, ...interface{}) {}

// searchesBy answers each query with the results listed for it.
func searchesBy(answers map[string][]SearchResult, asked *[]string) Searcher {
	return func(query string) ([]SearchResult, error) {
		*asked = append(*asked, query)
		return answers[query], nil
	}
}

func TestFindTrackPrefersTheOfficialVideo(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Blinding Lights The Weeknd official video": {
			{ID: "lyr", Title: "The Weeknd - Blinding Lights (Lyric Video)", Channel: "Fan"},
			{ID: "off", Title: "The Weeknd - Blinding Lights (Official Video)", Channel: "TheWeekndVEVO"},
		},
	}, &asked)

	got, err := FindTrack(search, "Blinding Lights", "The Weeknd", 0, quiet)
	if err != nil || got != (FoundTrack{VideoID: "off", Official: true}) {
		t.Fatalf("got %+v, %v; want the official video", got, err)
	}
	if len(asked) != 1 {
		t.Errorf("asked %v, want one search", asked)
	}
}

func TestFindTrackFallsBackToANonOfficialUpload(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Blinding Lights The Weeknd official video": {
			{ID: "cov", Title: "Blinding Lights (cover) The Weeknd", Channel: "Someone"},
		},
		"The Weeknd - Blinding Lights": {
			{ID: "aud", Title: "The Weeknd - Blinding Lights (Audio)", Channel: "The Weeknd - Topic"},
		},
	}, &asked)

	got, err := FindTrack(search, "Blinding Lights", "The Weeknd", 0, quiet)
	if err != nil || got != (FoundTrack{VideoID: "aud"}) {
		t.Fatalf("got %+v, %v; want the non-official audio upload", got, err)
	}
	if len(asked) != 3 {
		t.Errorf("asked %v, want every wording tried before settling for a non-official upload", asked)
	}
}

func TestFindTrackTakesTheOfficialAudioOfTheArtistsChannel(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Lovely Day Bill Withers official video": {
			{ID: "audio", Title: "Lovely Day (Official Audio)", Channel: "Bill Withers"},
		},
	}, &asked)

	got, err := FindTrack(search, "Lovely Day", "Bill Withers", 255, quiet)
	if err != nil || got != (FoundTrack{VideoID: "audio", Official: true}) {
		t.Fatalf("got %+v, %v; want the artist's official audio, counted as official", got, err)
	}
}

func TestFindTrackUsesTheLooseMatchOfAnyOfTheSearches(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Under Pressure Queen official video": {
			{ID: "lyr", Title: "Queen & David Bowie - Under Pressure (Lyrics)", Channel: "Lyrics Channel"},
		},
	}, &asked)

	got, err := FindTrack(search, "Under Pressure", "Queen, David Bowie", 0, quiet)
	if err != nil || got.VideoID != "lyr" || got.Official {
		t.Fatalf("got %+v, %v; want the lyric upload as a non-official match", got, err)
	}
	if len(asked) != 3 {
		t.Errorf("asked %v, want every wording tried", asked)
	}
}

func TestFindTrackAcceptsAnyOfSeveralArtists(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Under Pressure Queen official video": {
			{ID: "off", Title: "Under Pressure - David Bowie (Official Video)", Channel: "David Bowie"},
		},
	}, &asked)

	got, _ := FindTrack(search, "Under Pressure", "Queen, David Bowie", 0, quiet)
	if got != (FoundTrack{VideoID: "off", Official: true}) {
		t.Errorf("got %+v, want the video that names the second artist", got)
	}
}

func TestArtistNames(t *testing.T) {
	cases := map[string][]string{
		"The Weeknd":                {"The Weeknd"},
		"Queen, David Bowie":        {"Queen", "David Bowie"},
		"Drake & Don Toliver":       {"Drake & Don Toliver", "Drake", "Don Toliver"},
		"A, B & C":                  {"A", "B & C", "B", "C"},
		"Prince and the Revolution": {"Prince and the Revolution", "Prince", "the Revolution"},
		"Dr. Dre feat. Snoop Dogg":  {"Dr. Dre feat. Snoop Dogg", "Dr. Dre", "Snoop Dogg"},
		"Brandy":                    {"Brandy"},
		" ,  ":                      nil,
		"":                          nil,
	}
	for in, want := range cases {
		got := artistNames(in)
		if len(got) != len(want) {
			t.Errorf("artistNames(%q) = %q, want %q", in, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("artistNames(%q) = %q, want %q", in, got, want)
				break
			}
		}
	}
}

func TestFindTrackSearchesForTheWholeActAndMatchesEitherArtist(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Solar Eclipse Drake & Don Toliver official video": {
			{ID: "off", Title: "Don Toliver - Solar Eclipse (Official Video)", Channel: "Don Toliver"},
		},
	}, &asked)

	got, _ := FindTrack(search, "Solar Eclipse", "Drake & Don Toliver", 0, quiet)
	if got != (FoundTrack{VideoID: "off", Official: true}) || len(asked) != 1 {
		t.Errorf("got %+v after %v, want the video that names only one of the two artists", got, asked)
	}
}

func TestFindTrackFindsNothingWhenNothingMatches(t *testing.T) {
	var asked []string
	search := searchesBy(map[string][]SearchResult{
		"Blinding Lights The Weeknd official video": {{ID: "x", Title: "Totally different song", Channel: "Other"}},
		"The Weeknd - Blinding Lights":              {{ID: "y", Title: "Blinding Lights live karaoke The Weeknd", Channel: "Other"}},
	}, &asked)

	got, err := FindTrack(search, "Blinding Lights", "The Weeknd", 0, quiet)
	if err != nil || got.VideoID != "" {
		t.Errorf("got %+v, %v; want no video and no error", got, err)
	}
}

func TestFindTrackReportsASearchThatFails(t *testing.T) {
	failing := func(string) ([]SearchResult, error) { return nil, errors.New("offline") }
	if _, err := FindTrack(failing, "Song", "Artist", 0, quiet); err == nil {
		t.Error("want the search error")
	}
	if _, err := FindTrack(nil, "Song", "Artist", 0, quiet); err == nil {
		t.Error("want an error when there is no searcher")
	}
}
