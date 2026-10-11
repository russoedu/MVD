package official

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"youtube-downloader/libs/mvd-core/official/musicsource"
	"youtube-downloader/libs/mvd-core/official/trackfind"
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

// wikidataServer is a Wikidata client that asks a test server, which answers status and
// answer and remembers the query it was asked.
func wikidataServer(t *testing.T, status int, answer string, seen *string) *musicsource.WikidataClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = r.URL.Query().Get("query")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(server.Close)

	client := musicsource.NewWikidataClient()
	client.Endpoint = server.URL
	client.MinInterval = 0
	return client
}

func pickID(t *testing.T, results []trackpick.SearchResult, song trackpick.Song) (string, trackpick.Kind) {
	t.Helper()
	pick, ok := trackpick.PickBestVideo(results, song)
	if !ok {
		return "", trackpick.KindVideo
	}
	return pick.ID, pick.Kind
}

func TestFindBestVideoStopsAtAKnownVideoWithoutSearching(t *testing.T) {
	search := &scriptedSearch{}
	var music int
	sources := trackfind.Sources{
		Known: func(trackpick.Song, func(string, ...interface{})) []trackpick.SearchResult {
			return []trackpick.SearchResult{{ID: "known", Title: "Billy Joel - Uptown Girl (Official Video)", Channel: "billyjoelVEVO", MusicType: "OMV", Source: "Wikidata"}}
		},
		Music: func(string) ([]trackpick.SearchResult, error) { music++; return nil, nil },
	}

	pick, ok, _, err := trackfind.FindBestVideo(search.search, sources, trackpick.Song{Title: "Uptown Girl", Artists: []string{"Billy Joel"}}, quietLog)
	if err != nil || !ok || pick.ID != "known" || !strings.Contains(pick.Why, "listed by Wikidata") {
		t.Fatalf("got %+v, %v, %v", pick, ok, err)
	}
	if len(search.asked) != 0 || music != 0 {
		t.Errorf("a confident known video needs no search, asked %q and YouTube Music %d times", search.asked, music)
	}
}

func TestAKnownVideoIsTrustedWithoutTheWordOfficialOrTheArtistsChannel(t *testing.T) {
	song := trackpick.Song{Title: "Funkytown", Artists: []string{"Lipps Inc"}}
	known := []trackpick.SearchResult{{ID: "known", Title: "Lipps Inc - Funky Town", Channel: "TopPop", Source: "Wikidata"}}
	if id, _ := pickID(t, known, song); id != "known" {
		t.Errorf("got %q, want the video a database lists", id)
	}
	listedAudio := []trackpick.SearchResult{{ID: "audio", Title: "Lipps Inc - Funkytown (Audio)", Channel: "SomeFan", Source: "Wikidata"}}
	if id, _ := pickID(t, listedAudio, song); id != "" {
		t.Errorf("got %q, an audio upload still needs the artist's own channel", id)
	}
}

func TestFindBestVideoUsesYouTubeMusicWhenNothingIsKnown(t *testing.T) {
	search := &scriptedSearch{}
	var asked string
	sources := trackfind.Sources{
		Known: func(trackpick.Song, func(string, ...interface{})) []trackpick.SearchResult { return nil },
		Music: func(query string) ([]trackpick.SearchResult, error) {
			asked = query
			return []trackpick.SearchResult{
				{ID: "cover", Title: "Paradise Fears - MMMBop", Channel: "Paradise Fears", MusicType: "OMV", Views: 446000},
				{ID: "own", Title: "MMMBop (Official Music Video)", Channel: "Hanson", MusicType: "OMV", Views: 187000000},
			}, nil
		},
	}

	pick, ok, _, err := trackfind.FindBestVideo(search.search, sources, trackpick.Song{Title: "MMMBop", Artists: []string{"Hanson"}}, quietLog)
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
	search := &scriptedSearch{answers: map[string][]trackpick.SearchResult{
		"official video": {{ID: "official", Title: "Dr.Alban - It's My Life (Official 4K Video)", Channel: "Dr. Alban", Verified: true, Views: 93000000, Duration: 230}},
	}}
	sources := trackfind.Sources{
		Music: func(string) ([]trackpick.SearchResult, error) {
			return []trackpick.SearchResult{{ID: "totp", Title: "Dr. Alban - It's My Life (Top Of The Pops, 1st October, 1992)", Channel: "Dr. Alban", MusicType: "OMV", Views: 4000000, Duration: 189}}, nil
		},
	}
	song := trackpick.Song{Title: "It's My Life", Artists: []string{"Dr. Alban"}, DurationSec: 235}

	pick, ok, _, err := trackfind.FindBestVideo(search.search, sources, song, quietLog)
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
	search := &scriptedSearch{answers: map[string][]trackpick.SearchResult{
		"official video": {
			{ID: "arttrack1", Title: "Radio Edit - It's My Life", Channel: "Dr. Alban", Verified: true, Duration: 235, Views: 6000000},
			{ID: "official1", Title: "Dr.Alban - It's My Life (Official 4K Video)", Channel: "Dr. Alban", Verified: true, Duration: 230, Views: 93000000},
		},
	}}
	var asked []string
	sources := trackfind.Sources{
		Type: func(id string) (string, error) {
			asked = append(asked, id)
			if id == "arttrack1" {
				return "ATV", nil
			}
			return "OMV", nil
		},
	}
	song := trackpick.Song{Title: "It's My Life", Artists: []string{"Dr. Alban"}, DurationSec: 235}

	pick, ok, _, err := trackfind.FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || pick.ID != "official1" {
		t.Fatalf("got %+v, %v, %v; want the official video", pick, ok, err)
	}

	// Without a way to ask, the best candidate is taken as it is.
	pick, ok, _, _ = trackfind.FindBestVideo(search.search, trackfind.Sources{}, song, quietLog)
	if !ok || pick.ID == "" {
		t.Errorf("got %+v, %v; want a pick even when nobody can be asked", pick, ok)
	}
}

func TestFindBestVideoAsksOnlyAboutTheBestCandidates(t *testing.T) {
	var results []trackpick.SearchResult
	for i := 0; i < 12; i++ {
		results = append(results, trackpick.SearchResult{ID: "video" + string(rune('a'+i)), Title: "Band - Song", Channel: "Band", Views: int64(1000 + i)})
	}
	search := &scriptedSearch{answers: map[string][]trackpick.SearchResult{"official video": results}}
	asked := 0
	sources := trackfind.Sources{Type: func(string) (string, error) { asked++; return "ATV", nil }}

	if _, ok, _, _ := trackfind.FindBestVideo(search.search, sources, trackpick.Song{Title: "Song", Artists: []string{"Band"}}, quietLog); ok {
		t.Error("every candidate is an art track, so there is no pick")
	}
	if asked != trackfind.MaxChecked {
		t.Errorf("YouTube Music was asked %d times, want %d at most", asked, trackfind.MaxChecked)
	}
}

func TestFindBestVideoGoesOnWhenASourceFails(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]trackpick.SearchResult{
		"official video": {{ID: "official", Title: "Seal - Killer (Official Video)", Channel: "Seal", Verified: true, Views: 1400000}},
	}}
	sources := trackfind.Sources{
		Music: func(string) ([]trackpick.SearchResult, error) { return nil, http.ErrHandlerTimeout },
	}
	pick, ok, _, err := trackfind.FindBestVideo(search.search, sources, trackpick.Song{Title: "Killer", Artists: []string{"Seal"}}, quietLog)
	if err != nil || !ok || pick.ID != "official" {
		t.Errorf("a failing YouTube Music must not stop the search: got %+v, %v, %v", pick, ok, err)
	}
}

func TestFindBestVideoRemembersAndAsksNobodyTheSecondTime(t *testing.T) {
	cache := trackfind.NewResolutionCache(filepath.Join(t.TempDir(), "official.json"), time.Hour)
	search := &scriptedSearch{answers: map[string][]trackpick.SearchResult{
		"official video": {{ID: "official", Title: "Seal - Killer (Official Video)", Channel: "Seal", Verified: true, Views: 1400000}},
	}}
	song := trackpick.Song{Title: "Killer", Artists: []string{"Seal"}}
	sources := trackfind.Sources{Cache: cache}

	first, ok, _, err := trackfind.FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || first.ID != "official" {
		t.Fatalf("got %+v, %v, %v", first, ok, err)
	}
	asked := len(search.asked)

	second, ok, _, err := trackfind.FindBestVideo(search.search, sources, song, quietLog)
	if err != nil || !ok || second.ID != "official" || second.Kind != trackpick.KindVideo {
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
	res.TrackInfos = func(id string) (musicsource.TrackInfo, error) {
		if id == "ArtTrackTrk" {
			return musicsource.TrackInfo{Type: "ATV"}, nil
		}
		return musicsource.TrackInfo{Type: "OMV"}, nil
	}

	wikidata := wikidataServer(t, http.StatusOK, `{"results":{"bindings":[
	 {"vid":{"value":"VideoVideo1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"ArtTrackTrk"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"TopicTopic1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"OwnUploadI1"},"perfLabel":{"value":"Billy Joel"}},
	 {"vid":{"value":"GoneGoneGon"},"perfLabel":{"value":"Billy Joel"}}
	]}}`, nil)

	known := res.KnownCandidates(wikidata)(trackpick.Song{Title: "Uptown Girl", Artists: []string{"Billy Joel"}, OwnID: "OwnUploadI1"}, quietLog)
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
	if known := res.KnownCandidates(wikidata)(trackpick.Song{Title: "Song", Artists: []string{"Band"}}, quietLog); known != nil {
		t.Errorf("got %+v, want nothing when Wikidata is down", known)
	}
}
