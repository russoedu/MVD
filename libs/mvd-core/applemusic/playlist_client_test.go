package applemusic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const playlistLink = "https://music.apple.com/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb?l=pt"

func TestFetchReadsThePlaylistPage(t *testing.T) {
	var asked string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = r.URL.Path
		_, _ = w.Write([]byte(pageFixture))
	}))
	defer server.Close()

	c := NewClient()
	c.Rewrite = func(address string) string {
		return server.URL + strings.TrimPrefix(address, "https://music.apple.com")
	}
	got, err := c.Fetch(context.Background(), playlistLink)
	if err != nil {
		t.Fatal(err)
	}
	if asked != "/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb" || len(got.Tracks) != 2 {
		t.Errorf("asked %q, got %+v", asked, got)
	}
}

func TestFetchExplainsAMissingPlaylist(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	c := NewClient()
	c.Rewrite = func(string) string { return server.URL }
	_, err := c.Fetch(context.Background(), playlistLink)
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Errorf("want a hint that the playlist may be private, got %v", err)
	}
}

func TestFetchRefusesOtherLinks(t *testing.T) {
	if _, err := NewClient().Fetch(context.Background(), "https://example.com/x"); err == nil {
		t.Error("want an error for a link that is not a playlist")
	}
}
