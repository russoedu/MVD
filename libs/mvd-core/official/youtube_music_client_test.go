package official

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// musicServer answers the "next" endpoint with a document naming the video and its tag.
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

func TestVideoType(t *testing.T) {
	client := musicServer(t, map[string]string{
		"1okn3YNk-ro": `{"contents":{"videoId":"1okn3YNk-ro","musicVideoType":"MUSIC_VIDEO_TYPE_ATV"}}`,
		"NHozn0YXAeE": `{"contents":{"videoId":"NHozn0YXAeE","musicVideoType":"MUSIC_VIDEO_TYPE_OMV"}}`,
		"aaaaaaaaaaa": `{"contents":{"videoId":"aaaaaaaaaaa"}}`,
		"bbbbbbbbbbb": `{"contents":{"videoId":"someoneelse1","musicVideoType":"MUSIC_VIDEO_TYPE_UGC"}}`,
	})

	cases := map[string]string{
		"1okn3YNk-ro": "ATV",
		"NHozn0YXAeE": "OMV",
		"aaaaaaaaaaa": "", // no tag
		"bbbbbbbbbbb": "", // the answer is about another video
	}
	for id, want := range cases {
		got, err := client.VideoType(id)
		if err != nil || got != want {
			t.Errorf("VideoType(%q) = %q, %v; want %q", id, got, err, want)
		}
	}
}

func TestVideoTypeErrors(t *testing.T) {
	client := musicServer(t, nil)
	if _, err := client.VideoType("ccccccccccc"); err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("want the server's failure, got %v", err)
	}
	if _, err := client.VideoType("not an id"); err == nil {
		t.Error("an id that is not 11 characters of a video id must be refused before any request")
	}
}
