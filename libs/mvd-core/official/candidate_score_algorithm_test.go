package official

import "testing"

func pickID(t *testing.T, results []SearchResult, song Song) (string, Kind) {
	t.Helper()
	pick, ok := PickBestVideo(results, song)
	if !ok {
		return "", KindVideo
	}
	return pick.ID, pick.Kind
}

func TestPickBestVideoPrefersTheOfficialVideoOfTheArtistsChannel(t *testing.T) {
	results := []SearchResult{
		{ID: "topic", Title: "Wonderwall", Channel: "Oasis - Topic"},
		{ID: "lyrics", Title: "Oasis - Wonderwall (Lyrics)", Channel: "LyricsCo"},
		{ID: "cover", Title: "Wonderwall cover by Someone", Channel: "Someone"},
		{ID: "other", Title: "Wonderwall", Channel: "Random Guy"},
		{ID: "fan", Title: "Oasis - Wonderwall", Channel: "Fan Upload"},
		{ID: "official", Title: "Oasis - Wonderwall (Official Video)", Channel: "Oasis"},
	}
	song := Song{Title: "Wonderwall", Artists: []string{"Oasis"}, OwnID: "own"}

	if id, kind := pickID(t, results, song); id != "official" || kind != KindVideo {
		t.Errorf("got %q (%v), want the official video", id, kind)
	}
	if id, _ := pickID(t, results[:5], song); id != "" {
		t.Errorf("got %q, a fan upload that does not say official is not taken", id)
	}
	song.OwnID = "official"
	if id, _ := pickID(t, results, song); id == "official" {
		t.Error("must not pick the track's own video")
	}
}

func TestPickBestVideoTheCasesFromARealPlaylist(t *testing.T) {
	cases := []struct {
		name    string
		song    Song
		results []SearchResult
		want    string
	}{
		{
			"the video's title lacks the year of the track (Killer 2000 / Killer)",
			Song{Title: "Killer 2000", Artists: []string{"ATB"}, DurationSec: 336},
			[]SearchResult{
				{ID: "fan", Title: "ATB - Killer 2000 - HQ", Channel: "85KasiaD85"},
				{ID: "official", Title: "ATB - Killer", Channel: "ATB", Verified: true, Duration: 248, Views: 327066},
			},
			"official",
		},
		{
			"a split word (Run Away / Runaway), the artist's own channel beats a fan's official",
			Song{Title: "Run Away", Artists: []string{"Real McCoy"}},
			[]SearchResult{
				{ID: "fan", Title: "Real McCoy - Run Away (Official HD Video 1993)", Channel: "MusicBoxChannel"},
				{ID: "own", Title: "Real McCoy - Runaway (Official Video)", Channel: "Real McCoy"},
			},
			"own",
		},
		{
			"a Vevo channel whose title omits the artist",
			Song{Title: "Uptown Girl", Artists: []string{"Billy Joel"}},
			[]SearchResult{{ID: "vevo", Title: "Uptown Girl", Channel: "billyjoelVEVO"}},
			"vevo",
		},
		{
			"no 'official' in the title but the artist's own channel (How You Remind Me)",
			Song{Title: "How You Remind Me", Artists: []string{"Nickelback"}},
			[]SearchResult{
				{ID: "fan", Title: "Nickelback - How You Remind Me", Channel: "Fan Upload"},
				{ID: "own", Title: "Nickelback - How You Remind Me", Channel: "Nickelback", Verified: true},
			},
			"own",
		},
		{
			"a cover from the band's own channel is not the song",
			Song{Title: "MMMBop", Artists: []string{"Hanson"}},
			[]SearchResult{
				{ID: "cover", Title: "Paradise Fears - MMMBop (Hanson Cover)", Channel: "Paradise Fears"},
				{ID: "own", Title: "Hanson - MMMBop (Official Music Video)", Channel: "HANSON"},
			},
			"own",
		},
		{
			"another song of the same artist is not the song",
			Song{Title: "Alive And Kicking", Artists: []string{"Simple Minds"}},
			[]SearchResult{{ID: "other", Title: "Simple Minds - Don't You (Forget About Me) (Official Video)", Channel: "Simple Minds"}},
			"",
		},
		{
			"a live cut is rejected for a studio track and accepted for a live one",
			Song{Title: "Wonderwall", Artists: []string{"Oasis"}},
			[]SearchResult{{ID: "live", Title: "Oasis - Wonderwall (Live)", Channel: "Oasis"}},
			"",
		},
		{
			"two official videos: the one closer in length wins",
			Song{Title: "One", Artists: []string{"Metallica"}, DurationSec: 447},
			[]SearchResult{
				{ID: "short", Title: "Metallica - One (Official Video, edit)", Channel: "Metallica", Duration: 270},
				{ID: "full", Title: "Metallica: One (Official Music Video)", Channel: "Metallica", Duration: 465},
			},
			"full",
		},
	}
	for _, c := range cases {
		if id, _ := pickID(t, c.results, c.song); id != c.want {
			t.Errorf("%s: got %q, want %q", c.name, id, c.want)
		}
	}

	live := Song{Title: "Wonderwall (Live)", Artists: []string{"Oasis"}}
	results := []SearchResult{{ID: "live", Title: "Oasis - Wonderwall (Live)", Channel: "Oasis"}}
	if id, _ := pickID(t, results, live); id != "live" {
		t.Errorf("a live track accepts a live video, got %q", id)
	}
}

func TestPickBestVideoOfficialAudioIsTheFallbackNotTheFirstChoice(t *testing.T) {
	song := Song{Title: "Lovely Day", Artists: []string{"Bill Withers"}}
	audioOnly := []SearchResult{
		{ID: "fan", Title: "Bill Withers - Lovely Day (1978) (Remastered)", Channel: "Memory Lane"},
		{ID: "audio", Title: "Lovely Day (Official Audio)", Channel: "Bill Withers", Verified: true},
		{ID: "lyrics", Title: "Bill Withers - Lovely Day (Lyrics)", Channel: "LyricsCo"},
	}
	id, kind := pickID(t, audioOnly, song)
	if id != "audio" || kind != KindAudio {
		t.Errorf("got %q (%v), want the artist's official audio", id, kind)
	}

	withVideo := append(audioOnly, SearchResult{ID: "video", Title: "Bill Withers - Lovely Day (Official Video)", Channel: "Bill Withers"})
	if id, kind := pickID(t, withVideo, song); id != "video" || kind != KindVideo {
		t.Errorf("got %q (%v), a video beats an audio upload", id, kind)
	}

	// An official audio of someone else's channel is not the artist's.
	if id, _ := pickID(t, audioOnly[:1], song); id != "" {
		t.Errorf("got %q, a fan's upload is not taken", id)
	}
}

func TestPickBestVideoVerifiedChannelsCarryingTheArtistsNameAreTheirs(t *testing.T) {
	song := Song{Title: "Highway Star", Artists: []string{"Deep Purple"}}
	results := []SearchResult{
		{ID: "fan", Title: "Deep Purple - Highway Star (Official Video)", Channel: "SomeFan", Views: 900_000},
		{ID: "own", Title: "Highway Star", Channel: "Deep Purple Official", Verified: true, Views: 5_000_000},
	}
	if id, _ := pickID(t, results, song); id != "own" {
		t.Errorf("got %q, want the verified channel that carries the artist's name", id)
	}
}

func TestPickBestVideoDistrustsTinyReuploadsOfOtherChannels(t *testing.T) {
	song := Song{Title: "Run Away", Artists: []string{"Real McCoy"}}
	tiny := []SearchResult{{ID: "tiny", Title: "Real McCoy - Run Away (Official HD Video 1993)", Channel: "MusicBoxChannel", Views: 8_642}}
	if id, _ := pickID(t, tiny, song); id != "" {
		t.Errorf("got %q, a re-upload with 8,642 views is not trusted", id)
	}

	// The artist's own channel needs no audience, and a result with no view
	// count (the search did not say) is judged on the rest.
	own := []SearchResult{{ID: "own", Title: "Real McCoy - Run Away (Official Video)", Channel: "Real McCoy", Views: 300}}
	if id, _ := pickID(t, own, song); id != "own" {
		t.Errorf("got %q, want the artist's own upload whatever its views", id)
	}
	unknown := []SearchResult{{ID: "unknown", Title: "Real McCoy - Run Away (Official Video)", Channel: "SomeFan"}}
	if id, _ := pickID(t, unknown, song); id != "unknown" {
		t.Errorf("got %q, want a result without a view count to be judged on the rest", id)
	}
}

func TestRankCandidatesKeepsTheBestScoreOfAVideoSeenTwice(t *testing.T) {
	// YouTube Music names the official video by its metadata title, which says remix;
	// YouTube names it right. The video must not be lost to the first spelling.
	song := Song{Title: "It's My Life", Artists: []string{"Dr. Alban"}, DurationSec: 235}
	results := []SearchResult{
		{ID: "official", Title: "It's My Life (Raggadag Remix)", Channel: "Dr. Alban", MusicType: "OMV", Views: 93000000},
		{ID: "official", Title: "Dr.Alban - It's My Life (Official 4K Video)", Channel: "Dr. Alban", Verified: true, Duration: 230, Views: 93000000},
	}
	picks := RankCandidates(results, song)
	if len(picks) != 1 || picks[0].ID != "official" || !picks[0].Sure {
		t.Fatalf("got %+v, want the video once, with the score of its good title", picks)
	}
}

func TestDurationFit(t *testing.T) {
	cases := []struct {
		song, video int
		bonus       float64
	}{
		{200, 203, 1.5},
		{200, 215, 0.8},
		{200, 245, 0.2},
		{200, 400, -0.8},
		{0, 200, 0},
		{200, 0, 0},
	}
	for _, c := range cases {
		if _, bonus := durationFit(c.song, c.video); bonus != c.bonus {
			t.Errorf("durationFit(%d, %d) bonus = %v, want %v", c.song, c.video, bonus, c.bonus)
		}
	}
}
