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

func TestIdentifyTakesTheDatabaseNamesNotTheUploads(t *testing.T) {
	brainz := serving(t, musicBrainzAnswer, nil)
	identifier := &Identifier{MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL}}

	artist, title, ok := identifier.Identify("Tonight Is The Night (Le Click - Dance Mix )", "La Bouche")
	if !ok || artist != "Le Click" || title != "Tonight Is the Night" {
		t.Errorf("got %q, %q, %v", artist, title, ok)
	}
}

func TestIdentifyFallsBackToITunes(t *testing.T) {
	brainz := serving(t, `{"recordings":[]}`, nil)
	itunes := serving(t, iTunesAnswer, nil)
	identifier := &Identifier{
		MusicBrainz: &MusicBrainzClient{HTTP: brainz.Client(), Endpoint: brainz.URL},
		ITunes:      &ITunesClient{HTTP: itunes.Client(), Endpoint: itunes.URL},
	}

	artist, title, ok := identifier.Identify("Run to You", "Rage - Topic")
	if !ok || artist != "Rage" || title != "Run To You" {
		t.Errorf("got %q, %q, %v: the mix is not part of the title", artist, title, ok)
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
