package official

import (
	"errors"
	"strings"
	"testing"
)

// resolverWith returns a resolver whose YouTube pages say nothing (so only the search
// can find a video), with the given search and sources.
func resolverWith(t *testing.T, search *scriptedSearch, sources Sources) *Resolver {
	t.Helper()
	f := &fakeYouTube{}
	_, res := f.server(t)
	res.Searcher = search.search
	res.Sources = sources
	return res
}

func TestTheOfficialVideoIsSearchedForUnderTheNameADatabaseGives(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{
		"Le Click": {{ID: "official1", Title: "Le Click - Tonight Is The Night (Official Video)", Channel: "Le Click", Views: 3_000_000}},
	}}
	res := resolverWith(t, search, Sources{
		Identify: func(title, channel string) (string, string, bool) { return "Le Click", "Tonight Is The Night", true },
	})

	got, why := res.ResolveLog("up1", "Tonight Is The Night (Le Click - Dance Mix )", "La Bouche", 0, nil)
	if got != "official1" {
		t.Fatalf("got %q (%s), want the official video", got, why)
	}
	for _, query := range search.asked {
		if !strings.Contains(query, "Le Click") {
			t.Errorf("searched for %q, which leaves the artist out", query)
		}
	}
}

func TestAPlainVideoIsLeftAloneUnlessADatabaseNamesIt(t *testing.T) {
	search := &scriptedSearch{}
	res := resolverWith(t, search, Sources{})
	res.TrackInfos = func(string) (TrackInfo, error) { return TrackInfo{Type: "UGC"}, nil }

	if got, _ := res.ResolveLog("up1", "ICE MC - Cinema (Radio Edit) (1990)", "Some Channel", 0, nil); got != "" {
		t.Errorf("got %q, a plain video no database knows is kept", got)
	}
	if len(search.asked) != 0 {
		t.Errorf("searched for %q", search.asked)
	}

	res.Sources.Identify = func(string, string) (string, string, bool) { return "ICE MC", "Cinema", true }
	search.answers = map[string][]SearchResult{
		"ICE MC": {{ID: "official2", Title: "ICE MC - Cinema (Official Video)", Channel: "ICE MC", Views: 1_000_000}},
	}
	if got, _ := res.ResolveLog("up1", "ICE MC - Cinema (Radio Edit) (1990)", "Some Channel", 0, nil); got != "official2" {
		t.Errorf("got %q, a plain video a database names is searched for", got)
	}
}

func fanUploads() []SearchResult {
	return []SearchResult{
		{ID: "fanA", Title: "Rage - Run To You", Channel: "ohnoitisnathan44", Views: 5_000},
		{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000},
		{ID: "cover", Title: "Rage - Run To You (cover)", Channel: "someone", Views: 900_000},
	}
}

func TestWithoutAnOfficialVideoTheBestQualityUploadIsTaken(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": fanUploads()}}
	asked := map[string]bool{}
	qualities := map[string]Quality{"up1": {AudioKbps: 128}, "fanA": {Height: 360, AudioKbps: 96}, "fanB": {Height: 1080, AudioKbps: 160}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality: func(id string) (Quality, error) {
			asked[id] = true
			return qualities[id], nil
		},
	})

	got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil)
	if got.ID != "fanB" || got.Official || !strings.Contains(got.Reason, "best quality") {
		t.Errorf("got %+v, want the 1080p upload, not as an official video", got)
	}
	if id, _ := res.ResolveLog("up1", "Run to You", "Rage - Topic", 0, nil); id != "" {
		t.Errorf("got %q: ResolveLog answers with official videos only", id)
	}
	if asked["cover"] {
		t.Error("a cover is no upload of the song, its quality is not worth looking up")
	}
}

func TestThePlaylistsOwnUploadStaysWhenNothingIsBetter(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": fanUploads()}}
	qualities := map[string]Quality{"up1": {Height: 1080, AudioKbps: 160}, "fanA": {Height: 360, AudioKbps: 96}, "fanB": {Height: 1080, AudioKbps: 160}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
	})
	res.TrackInfos = func(string) (TrackInfo, error) { return TrackInfo{Type: "UGC"}, nil }

	if got := res.ResolveVersion("up1", "Rage - Run to You", "Some Channel", 0, nil); got.ID != "" {
		t.Errorf("got %+v, an upload only as good as the playlist's own is no improvement", got)
	}
}

func TestAnArtTracksStillPictureDoesNotCountAsAPicture(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": fanUploads()}}
	qualities := map[string]Quality{"up1": {Height: 1080, AudioKbps: 130}, "fanA": {Height: 480, AudioKbps: 130}, "fanB": {Height: 1080, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want a real video over the art track's 1080p still image", got)
	}
}

func TestWhenQualityCannotBeLookedUpTheOriginalStays(t *testing.T) {
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": fanUploads()}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(string) (Quality, error) { return Quality{}, errors.New("blocked") },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "" {
		t.Errorf("got %+v", got)
	}
}

func TestQualityScoreCountsThePictureMoreThanTheSound(t *testing.T) {
	if (Quality{Height: 1080}).Score() <= (Quality{AudioKbps: 320}).Score() {
		t.Error("a 1080p picture should beat the best audio alone")
	}
	if (Quality{Height: 4320, AudioKbps: 999}).Score() > 1 {
		t.Error("more than the ceilings does not count")
	}
	if (Quality{}).Score() != 0 {
		t.Error("nothing scores nothing")
	}
}

func TestAnArtTrackUnderTheArtistsOwnNameIsNotAnUploadToTake(t *testing.T) {
	uploads := []SearchResult{
		{ID: "artTrack", Title: "Rage - Run To You", Channel: "Rage", Views: 800_000},
		{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000},
	}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {AudioKbps: 130}, "artTrack": {Height: 1080, AudioKbps: 140}, "fanB": {Height: 720, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
		Type: func(id string) (string, error) {
			if id == "artTrack" {
				return "ATV", nil
			}
			return "UGC", nil
		},
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want the real video, not another art track", got)
	}
}

func TestTelevisionAppearancesAndFanEditsAreNotTheSongsVideo(t *testing.T) {
	uploads := []SearchResult{
		{ID: "tv", Title: "Rage no Raul Gil (1995) Run To You & Interview", Channel: "XiaoAn", Views: 800_000},
		{ID: "edit", Title: "Rage - Run To You (Extended Intro Cut - Tony Mendes Video Re Edit)", Channel: "Tony Mendes", Views: 700_000},
		{ID: "totp", Title: "Rage - Run To You - TOTP HQ", Channel: "gigantis2000", Views: 600_000},
		{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2001", Views: 90_000},
	}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {AudioKbps: 130}, "tv": {Height: 1080, AudioKbps: 140}, "edit": {Height: 1080, AudioKbps: 140}, "totp": {Height: 1080, AudioKbps: 140}, "fanB": {Height: 720, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want the plain video", got)
	}
}

func TestLyricAndAudioUploadsAreNotTheSongsVideo(t *testing.T) {
	uploads := []SearchResult{
		{ID: "lyrics", Title: "Rage - Run To You (Letra/Lyrics)", Channel: "LyricsFan", Views: 800_000},
		{ID: "audio", Title: "Rage - Run To You (Audio)", Channel: "AudioFan", Views: 700_000},
		{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000},
	}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {AudioKbps: 130}, "lyrics": {Height: 1080, AudioKbps: 140}, "audio": {Height: 1080, AudioKbps: 140}, "fanB": {Height: 720, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want the plain video", got)
	}
}

func TestAnUploadThatIsOnlyAStillPictureIsLeftOut(t *testing.T) {
	uploads := []SearchResult{
		{ID: "cover", Title: "Rage - Run To You (HQ)", Channel: "CoverFan", Views: 800_000},
		{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000},
	}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {AudioKbps: 130}, "cover": {Height: 1080, AudioKbps: 160}, "fanB": {Height: 720, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
		Still:    func(id string) (bool, error) { return id == "cover", nil },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want the video that moves over the still picture of better size", got)
	}
}

func TestThePlaylistsOwnStillPictureCountsForNoPicture(t *testing.T) {
	uploads := []SearchResult{{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000}}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {Height: 1080, AudioKbps: 130}, "fanB": {Height: 480, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
		Still:    func(id string) (bool, error) { return id == "up1", nil },
	})
	// The playlist's own upload is not an art track, but a plain video that is only a picture.
	res.TrackInfos = func(string) (TrackInfo, error) { return TrackInfo{Type: "UGC"}, nil }

	if got := res.ResolveVersion("up1", "Rage - Run to You", "Some Channel", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v, want the 480p video over a 1080p still picture", got)
	}
}

func TestWhenStillnessCannotBeToldNothingIsLeftOut(t *testing.T) {
	uploads := []SearchResult{{ID: "fanB", Title: "Rage - Run To You (HD)", Channel: "gigantis2000", Views: 90_000}}
	search := &scriptedSearch{answers: map[string][]SearchResult{"Rage": uploads}}
	qualities := map[string]Quality{"up1": {AudioKbps: 130}, "fanB": {Height: 720, AudioKbps: 131}}
	res := resolverWith(t, search, Sources{
		Identify: func(string, string) (string, string, bool) { return "Rage", "Run To You", true },
		Quality:  func(id string) (Quality, error) { return qualities[id], nil },
		Still:    func(string) (bool, error) { return false, errors.New("image server down") },
	})

	if got := res.ResolveVersion("up1", "Run to You", "Rage - Topic", 0, nil); got.ID != "fanB" {
		t.Errorf("got %+v: a failing image server must not cost a good version", got)
	}
}
