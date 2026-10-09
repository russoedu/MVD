package official

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// panelAnswer is a "next" answer shaped like YouTube Music's: the panel entry
// of the video, with its title, byline, length and music video tag.
func panelAnswer(id, tag, title, artist, byline, length string) string {
	return `{"contents":{"watch":{"tabs":[{"playlist":{"items":[{"playlistPanelVideoRenderer":{` +
		`"title":{"runs":[{"text":"` + title + `"}]},` +
		`"shortBylineText":{"runs":[{"text":"` + artist + `"}]},` +
		`"longBylineText":{"runs":[{"text":"` + byline + `"}]},` +
		`"lengthText":{"runs":[{"text":"` + length + `"}]},` +
		`"videoId":"` + id + `",` +
		`"navigationEndpoint":{"watchEndpoint":{"watchEndpointMusicSupportedConfigs":{"musicVideoType":"` + tag + `"}}}}}]}}]}}}`
}

// musicServer answers the "next" endpoint with a document per video id.
func musicServer(t *testing.T, answers map[string]string) *YouTubeMusicClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			VideoID string `json:"videoId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Header.Get("Origin") != "https://music.youtube.com" {
			t.Errorf("the request must look like YouTube Music's own, Origin = %q", r.Header.Get("Origin"))
		}
		answer, ok := answers[body.VideoID]
		if !ok {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(server.Close)

	client := NewYouTubeMusicClient()
	client.NextURL = server.URL
	return client
}

func TestDescribe(t *testing.T) {
	client := musicServer(t, map[string]string{
		"8rIjsa85UVk": panelAnswer("8rIjsa85UVk", "MUSIC_VIDEO_TYPE_ATV", "Running Up That Hill (A Deal With God)", "Kate Bush", "Kate Bush • Hounds Of Love • 1985", "4:59"),
		"NHozn0YXAeE": panelAnswer("NHozn0YXAeE", "MUSIC_VIDEO_TYPE_OMV", "MMMBop (Official Music Video)", "Hanson", "Hanson • 187M views", "3:52"),
		"aaaaaaaaaaa": `{"contents":{"watch":{"tabs":[{"playlist":{"items":[{"playlistPanelVideoRenderer":{"videoId":"aaaaaaaaaaa"}}]}}]}}}`,
		"bbbbbbbbbbb": panelAnswer("someoneelse1", "MUSIC_VIDEO_TYPE_UGC", "Other", "Other", "Other • x", "1:00"),
		"ccccccccccc": panelAnswer("ccccccccccc", "MUSIC_VIDEO_TYPE_ATV", "Long", "Band", "Band", "1:02:03"),
	})

	got, err := client.Describe("8rIjsa85UVk")
	want := TrackInfo{Type: "ATV", Title: "Running Up That Hill (A Deal With God)", Artist: "Kate Bush", Album: "Hounds Of Love", DurationSec: 299}
	if err != nil || got != want {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}

	if got, err := client.Describe("NHozn0YXAeE"); err != nil || got.Type != "OMV" || got.Artist != "Hanson" || got.Album != "" {
		t.Errorf("an official video: got %+v, %v (a view count is not an album)", got, err)
	}
	if got, err := client.Describe("aaaaaaaaaaa"); err != nil || got != (TrackInfo{}) {
		t.Errorf("no tag and no text: got %+v, %v", got, err)
	}
	if got, err := client.Describe("bbbbbbbbbbb"); err != nil || got != (TrackInfo{}) {
		t.Errorf("an answer about another video says nothing: got %+v, %v", got, err)
	}
	if got, _ := client.Describe("ccccccccccc"); got.DurationSec != 3723 {
		t.Errorf("hours, minutes and seconds: got %d, want 3723", got.DurationSec)
	}
}

func TestDescribeErrors(t *testing.T) {
	client := musicServer(t, nil)
	if _, err := client.Describe("ddddddddddd"); err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("want the server's failure, got %v", err)
	}
	if _, err := client.Describe("not an id"); err == nil {
		t.Error("an id that is not 11 characters of a video id must be refused before any request")
	}
}

func TestParseClock(t *testing.T) {
	cases := map[string]int{"4:59": 299, "0:05": 5, "1:02:03": 3723, "": 0, "abc": 0, "1:xx": 0}
	for in, want := range cases {
		if got := parseClock(in); got != want {
			t.Errorf("parseClock(%q) = %d, want %d", in, got, want)
		}
	}
}
