package official

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindBestVideoStopsAtAKnownVideoWithoutSearching(t *testing.T) {
	search := &scriptedSearch{}
	var music int
	sources := Sources{
		Known: func(Song, func(string, ...interface{})) []SearchResult {
			return []SearchResult{{ID: "known", Title: "Billy Joel - Uptown Girl (Official Video)", Channel: "billyjoelVEVO", MusicType: "OMV", Source: "Wikidata"}}
		},
		Music: func(string) ([]SearchResult, error) { music++; return nil, nil },
	}

	pick, ok, _, err := FindBestVideo(search.search, sources, Song{Title: "Uptown Girl", Artists: []string{"Billy Joel"}}, quietLog)
	if err != nil || !ok || pick.ID != "known" || !strings.Contains(pick.Why, "listed by Wikidata") {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if len(search.asked) != 0 || music != 0 {
		t.Errorf("a confident known video needs no search, asked %q and YouTube Music %d times", search.asked, music)
	}
}

func TestAKnownVideoIsTrustedWithoutTheWordOfficialOrTheArtistsChannel(t *testing.T) {
	song := Song{Title: "Funkytown", Artists: []string{"Lipps Inc"}}
	known := []SearchResult{{ID: "known", Title: "Lipps Inc - Funky Town", Channel: "TopPop", Source: "Wikidata"}}
	if id, _ := pickID(t, known, song); id != "known" {
		t.Errorf("got %q, want the video a database lists", id)
	}
	listedAudio := []SearchResult{{ID: "audio", Title: "Lipps Inc - Funkytown (Audio)", Channel: "SomeFan", Source: "Wikidata"}}
	if id, _ := pickID(t, listedAudio, song); id != "" {
		t.Errorf("got %q, an audio upload still needs the artist's own channel", id)
	}
}

func TestFindBestVideoUsesYouTubeMusicWhenNothingIsKnown(t *testing.T) {
	search := &scriptedSearch{}
	var asked string
	sources := Sources{
		Known: func(Song, func(string, ...interface{})) []SearchResult { return nil },
		Music: func(query string) ([]SearchResult, error) {
			asked = query
			return []SearchResult{
				{ID: "cover", Title: "Paradise Fears - MMMBop", Channel: "Paradise Fears", MusicType: "OMV", Views: 446000},
				{ID: "own", Title: "MMMBop (Official Music Video)", Channel: "Hanson", MusicType: "OMV", Views: 187000000},
			}, nil
		},
	}

	pick, ok, _, err := FindBestVideo(search.search, sources, Song{Title: "MMMBop", Artists: []string{"Hanson"}}, quietLog)
	if err != nil || !ok || pick.ID != "own" {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if asked != "MMMBop Hanson" || len(search.asked) != 0 {
		t.Errorf("YouTube Music was asked %q and YouTube %q; want only YouTube Music, for the title and artist", asked, search.asked)
	}
}

func TestFindBestVideoDoesNotStopAtAnArtistUploadThatSaysNothing(t *testing.T) {
	// Dr. Alban's channel has the official video and a TV performance of the same
	// song; YouTube Music's search tags both as official videos.
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official video": {{ID: "official", Title: "Dr.Alban - It's My Life (Official 4K Video)", Channel: "Dr. Alban", Verified: true, Views: 93000000, Duration: 230}},
	}}
	sources := Sources{
		Music: func(string) ([]SearchResult, error) {
			return []SearchResult{{ID: "totp", Title: "Dr. Alban - It's My Life (Top Of The Pops, 1st October, 1992)", Channel: "Dr. Alban", MusicType: "OMV", Views: 4000000, Duration: 189}}, nil
		},
	}
	song := Song{Title: "It's My Life", Artists: []string{"Dr. Alban"}, DurationSec: 235}

	pick, ok, _, err := FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || pick.ID != "official" {
		t.Fatalf("got %+v, %v, %v; want the official video, not the performance", pick, ok, err)
	}
	if len(search.asked) == 0 {
		t.Error("the search must go on past a video that does not say it is official")
	}
}

func TestFindBestVideoSkipsAnArtTrackThatLooksLikeTheArtistsVideo(t *testing.T) {
	// YouTube shows an art track under the artist's name: verified, same length, no
	// "official" in its title. YouTube Music knows it for what it is.
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official video": {
			{ID: "arttrack1", Title: "Radio Edit - It's My Life", Channel: "Dr. Alban", Verified: true, Duration: 235, Views: 6000000},
			{ID: "official1", Title: "Dr.Alban - It's My Life (Official 4K Video)", Channel: "Dr. Alban", Verified: true, Duration: 230, Views: 93000000},
		},
	}}
	var asked []string
	sources := Sources{
		Type: func(id string) (string, error) {
			asked = append(asked, id)
			if id == "arttrack1" {
				return "ATV", nil
			}
			return "OMV", nil
		},
	}
	song := Song{Title: "It's My Life", Artists: []string{"Dr. Alban"}, DurationSec: 235}

	pick, ok, _, err := FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || pick.ID != "official1" {
		t.Fatalf("got %+v, %v, %v; want the official video", pick, ok, err)
	}

	// Without a way to ask, the best candidate is taken as it is.
	pick, ok, _, _ = FindBestVideo(search.search, Sources{}, song, quietLog)
	if !ok || pick.ID == "" {
		t.Errorf("got %+v, %v; want a pick even when nobody can be asked", pick, ok)
	}
}

func TestFindBestVideoAsksOnlyAboutTheBestCandidates(t *testing.T) {
	var results []SearchResult
	for i := 0; i < 12; i++ {
		results = append(results, SearchResult{ID: "video" + string(rune('a'+i)), Title: "Band - Song", Channel: "Band", Views: int64(1000 + i)})
	}
	search := &scriptedSearch{answers: map[string][]SearchResult{"official video": results}}
	asked := 0
	sources := Sources{Type: func(string) (string, error) { asked++; return "ATV", nil }}

	if _, ok, _, _ := FindBestVideo(search.search, sources, Song{Title: "Song", Artists: []string{"Band"}}, quietLog); ok {
		t.Error("every candidate is an art track, so there is no pick")
	}
	if asked != maxChecked {
		t.Errorf("YouTube Music was asked %d times, want %d at most", asked, maxChecked)
	}
}

func TestFindBestVideoGoesOnWhenASourceFails(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official video": {{ID: "official", Title: "Seal - Killer (Official Video)", Channel: "Seal", Verified: true, Views: 1400000}},
	}}
	sources := Sources{
		Music: func(string) ([]SearchResult, error) { return nil, http.ErrHandlerTimeout },
	}
	pick, ok, _, err := FindBestVideo(search.search, sources, Song{Title: "Killer", Artists: []string{"Seal"}}, quietLog)
	if err != nil || !ok || pick.ID != "official" {
		t.Errorf("a failing YouTube Music must not stop the search: got %+v, %v, %v", pick, ok, err)
	}
}

func TestFindBestVideoRemembersAndAsksNobodyTheSecondTime(t *testing.T) {
	cache := NewResolutionCache(filepath.Join(t.TempDir(), "official.json"), time.Hour)
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"official video": {{ID: "official", Title: "Seal - Killer (Official Video)", Channel: "Seal", Verified: true, Views: 1400000}},
	}}
	song := Song{Title: "Killer", Artists: []string{"Seal"}}
	sources := Sources{Cache: cache}

	first, ok, _, err := FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || first.ID != "official" {
		t.Fatalf("got %+v, %v, %v", first, ok, err)
	}
	asked := len(search.asked)

	second, ok, _, err := FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || second.ID != "official" || second.Kind != KindVideo {
		t.Fatalf("got %+v, %v, %v", second, ok, err)
	}
	if len(search.asked) != asked {
		t.Errorf("the second run searched again: %q", search.asked)
	}
	if !strings.Contains(second.Why, "remembered") {
		t.Errorf("the log should say it was remembered, got %q", second.Why)
	}
}

func TestKnownCandidatesKeepsRealVideosOnly(t *testing.T) {
	f := &fakeYouTube{
		authors: map[string]string{"VideoVideo1": "billyjoelVEVO", "ArtTrackTrk": "Billy Joel", "TopicTopic1": "Billy Joel - Topic", "OwnUploadI1": "Billy Joel"},
		titles:  map[string]string{"VideoVideo1": "Billy Joel - Uptown Girl (Official Video)", "ArtTrackTrk": "Uptown Girl"},
	}
	_, res := f.server(t)
	res.TrackInfos = func(id string) (TrackInfo, error) {
		if id == "ArtTrackTrk" {
			return TrackInfo{Type: "ATV"}, nil
		}
		return TrackInfo{Type: "OMV"}, nil
	}

	wikidata := wikidataServer(t, http.StatusOK, `{"results":{"bindings":[
	 {"vid":{"value":"VideoVideo1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"ArtTrackTrk"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"TopicTopic1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"OwnUploadI1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"GoneGoneGon"},"perfLabel":{"value":"Billy Joel"}}
	]}}`, nil)

	known := res.KnownCandidates(wikidata)(Song{Title: "Uptown Girl", Artists: []string{"Billy Joel"}, OwnID: "OwnUploadI1"}, quietLog)
	if len(known) != 1 {
		t.Fatalf("got %+v, want only the real video: the art track, the Topic upload, the song's own upload and a deleted video are out", known)
	}
	if known[0].ID != "VideoVideo1" || known[0].Source != "Wikidata" || known[0].Channel != "billyjoelVEVO" || known[0].MusicType != "OMV" || !strings.Contains(known[0].Title, "Uptown Girl") {
		t.Errorf("got %+v", known[0])
	}
}

func TestKnownCandidatesSurvivesAWikidataFailure(t *testing.T) {
	f := &fakeYouTube{}
	_, res := f.server(t)
	wikidata := wikidataServer(t, http.StatusServiceUnavailable, "down", nil)
	if known := res.KnownCandidates(wikidata)(Song{Title: "Song", Artists: []string{"Band"}}, quietLog); known != nil {
		t.Errorf("got %+v, want nothing when Wikidata is down", known)
	}
}

func TestSearchVideosReadsYouTubeMusicRows(t *testing.T) {
	answer := `{"contents":{"sectionListRenderer":{"contents":[{"musicShelfRenderer":{"contents":[
	 {"musicResponsiveListItemRenderer":{"flexColumns":[
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"Killer"}]}}},
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"ATB"},{"text":" • "},{"text":"327K views"},{"text":" • "},{"text":"4:08"}]}}}],
	  "overlay":{"navigationEndpoint":{"watchEndpoint":{"videoId":"sJ0FFbRBQlE","watchEndpointMusicSupportedConfigs":{"musicVideoType":"MUSIC_VIDEO_TYPE_OMV"}}}}}},
	 {"musicResponsiveListItemRenderer":{"flexColumns":[
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"ATB - Killer 2000 - HQ"}]}}},
	   {"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"85KasiaD85"},{"text":" • "},{"text":"1.2M views"},{"text":" • "},{"text":"5:36"}]}}}],
	  "overlay":{"navigationEndpoint":{"watchEndpoint":{"videoId":"1eoPsEBErws","watchEndpointMusicSupportedConfigs":{"musicVideoType":"MUSIC_VIDEO_TYPE_UGC"}}}}}},
	 {"musicResponsiveListItemRenderer":{"flexColumns":[{"musicResponsiveListItemFlexColumnRenderer":{"text":{"runs":[{"text":"no id here"}]}}}]}}
	]}}]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "https://music.youtube.com" {
			t.Errorf("the request must look like YouTube Music's own")
		}
		_, _ = w.Write([]byte(answer))
	}))
	defer server.Close()

	client := NewYouTubeMusicClient()
	client.SearchURL = server.URL
	got, err := client.SearchVideos("Killer ATB")
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v; want the two rows that name a video", got, err)
	}
	want := SearchResult{ID: "sJ0FFbRBQlE", Title: "Killer", Channel: "ATB", Views: 327000, Duration: 248, MusicType: "OMV"}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
	if got[1].MusicType != "UGC" || got[1].Views != 1200000 || got[1].Duration != 336 {
		t.Errorf("got %+v", got[1])
	}
}

func TestParseCount(t *testing.T) {
	cases := map[string]int64{"327K views": 327000, "1.2M views": 1200000, "540 views": 540, "2.1B views": 2100000000, "1,234 views": 1234, "": 0, "no views": 0}
	for in, want := range cases {
		if got := parseCount(in); got != want {
			t.Errorf("parseCount(%q) = %d, want %d", in, got, want)
		}
	}
}
