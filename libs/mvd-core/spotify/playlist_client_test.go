package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchReadsTheEmbedPage(t *testing.T) {
	var asked string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = r.URL.Path
		_, _ = w.Write([]byte(embedFixture))
	}))
	defer server.Close()

	c := NewClient()
	c.EmbedBase = server.URL + "/embed/playlist/"
	got, err := c.Fetch(context.Background(), "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=x")
	if err != nil {
		t.Fatal(err)
	}
	if asked != "/embed/playlist/37i9dQZF1DXcBWIGoYBM5M" || len(got.Tracks) != 2 {
		t.Errorf("asked %q, got %+v", asked, got)
	}
}

func TestFetchExplainsAMissingPlaylist(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	c := NewClient()
	c.EmbedBase = server.URL + "/"
	_, err := c.Fetch(context.Background(), "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M")
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Errorf("want a hint that the playlist may be private, got %v", err)
	}
}

func TestFetchRefusesOtherLinks(t *testing.T) {
	if _, err := NewClient().Fetch(context.Background(), "https://example.com/x"); err == nil {
		t.Error("want an error for a link that is not a playlist")
	}
}
