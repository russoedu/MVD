package songid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const musicBrainzAnswer = `{"recordings":[
  {"title":"Tonight Is the Night","score":100,"artist-credit":[{"name":"Le Click"}]},
  {"title":"Tonight Is the Night","score":40,"artist-credit":[{"name":"Someone Else"}]}]}`

const iTunesAnswer = `{"results":[{"trackName":"Run To You (Radio Edit)","artistName":"Rage"}]}`

func serving(t *testing.T, body string, asked *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if asked != nil {
			*asked = append(*asked, r.URL.Query().Get("query")+r.URL.Query().Get("term"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestMusicBrainzAnswersAboveTheScoreAreKept(t *testing.T) {
	var asked []string
	server := serving(t, musicBrainzAnswer, &asked)
	client := &MusicBrainzClient{HTTP: server.Client(), Endpoint: server.URL, UserAgent: "test"}

	got, err := client.Recordings("Le Click", `Tonight "Is" The Night`)
	if err != nil || len(got) != 1 || got[0] != (Identity{"Le Click", "Tonight Is the Night"}) {
		t.Fatalf("got %+v, %v", got, err)
	}
	if len(asked) != 1 || !strings.Contains(asked[0], `recording:"Tonight \"Is\" The Night" AND artist:"Le Click"`) {
		t.Errorf("asked %q", asked)
	}
}

func TestMusicBrainzKeepsItsDistance(t *testing.T) {
	server := serving(t, `{"recordings":[]}`, nil)
	client := &MusicBrainzClient{HTTP: server.Client(), Endpoint: server.URL, MinInterval: 60 * time.Millisecond}

	start := time.Now()
	_, _ = client.Recordings("a", "b")
	_, _ = client.Recordings("a", "b")
	if time.Since(start) < 55*time.Millisecond {
		t.Error("two requests came too close together")
	}
}

func TestMusicBrainzErrorsAreReported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer server.Close()
	client := &MusicBrainzClient{HTTP: server.Client(), Endpoint: server.URL}
	if _, err := client.Recordings("a", "b"); err == nil {
		t.Error("a 503 is an error")
	}
}

func TestIdentifyAsksITunesFirstAndTakesItsNames(t *testing.T) {
	var brainzAsked []string
	brainz := serving(t, musicBrainzAnswer, &brainzAsked)
	itunes := serving(t, iTunesAnswer, nil)
	identifier := &Identifier{
		MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL},
		ITunes:      &ITunesClient{HTTP: itunes.Client(), Endpoint: itunes.URL},
	}

	artist, title, ok := identifier.Identify("Run to You", "Rage - Topic")
	if !ok || artist != "Rage" || title != "Run To You" {
		t.Errorf("got %q, %q, %v: the mix is not part of the title", artist, title, ok)
	}
	if len(brainzAsked) != 0 {
		t.Errorf("MusicBrainz was asked %q although iTunes knew the song", brainzAsked)
	}
}

func TestIdentifyFallsBackToMusicBrainzForWhatITunesDoesNotKnow(t *testing.T) {
	brainz := serving(t, musicBrainzAnswer, nil)
	itunes := serving(t, `{"results":[]}`, nil)
	identifier := &Identifier{
		MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL},
		ITunes:      &ITunesClient{HTTP: itunes.Client(), Endpoint: itunes.URL},
	}

	artist, title, ok := identifier.Identify("Tonight Is The Night (Le Click - Dance Mix )", "La Bouche")
	if !ok || artist != "Le Click" || title != "Tonight Is the Night" {
		t.Errorf("got %q, %q, %v", artist, title, ok)
	}
}

func TestIdentifyWorksWithMusicBrainzAlone(t *testing.T) {
	brainz := serving(t, musicBrainzAnswer, nil)
	identifier := &Identifier{MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL}}

	if artist, _, ok := identifier.Identify("Tonight Is The Night (Le Click - Dance Mix )", "La Bouche"); !ok || artist != "Le Click" {
		t.Errorf("got %q, %v", artist, ok)
	}
}

func TestIdentifyRefusesAnAnswerTheUploadDoesNotMention(t *testing.T) {
	itunes := serving(t, `{"results":[{"trackName":"Run To You","artistName":"Bryan Adams"}]}`, nil)
	identifier := &Identifier{ITunes: &ITunesClient{HTTP: itunes.Client(), Endpoint: itunes.URL}}

	if _, _, ok := identifier.Identify("Run to You", "Rage - Topic"); ok {
		t.Error("another artist's song of the same name is not an answer")
	}
}

func TestIdentifyWithoutDatabasesFindsNothing(t *testing.T) {
	if _, _, ok := (&Identifier{}).Identify("A - B", "c"); ok {
		t.Error("no database, no answer")
	}
}

const deezerAnswer = `{"data":[
  {"title":"I Drove All Night","artist":{"name":"Bandit"}},
  {"title":"I Drove All Night (Club Mix)","artist":{"name":"Bandido"}}]}`

func TestDeezerTracksAreReadWithTheirArtists(t *testing.T) {
	server := serving(t, deezerAnswer, nil)
	client := &DeezerClient{HTTP: server.Client(), Endpoint: server.URL}

	got, err := client.Tracks("Bandido I Drove All Night")
	if err != nil || len(got) != 2 || got[1].Artist != "Bandido" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestIdentifyAsksDeezerBeforeMusicBrainzAndSkipsAnotherArtistsSong(t *testing.T) {
	var brainzAsked []string
	brainz := serving(t, musicBrainzAnswer, &brainzAsked)
	itunes := serving(t, `{"results":[]}`, nil)
	deezer := serving(t, deezerAnswer, nil)
	identifier := &Identifier{
		MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL},
		ITunes:      &ITunesClient{HTTP: itunes.Client(), Endpoint: itunes.URL},
		Deezer:      &DeezerClient{HTTP: deezer.Client(), Endpoint: deezer.URL},
	}

	artist, _, ok := identifier.Identify("Bandido - I Drove All Night (1991)", "Eurodance")
	if !ok || artist != "Bandido" {
		t.Errorf("got %q, %v: the first answer is another artist's song of the same name", artist, ok)
	}
	if len(brainzAsked) != 0 {
		t.Errorf("MusicBrainz was asked %q although Deezer knew the song", brainzAsked)
	}
}
