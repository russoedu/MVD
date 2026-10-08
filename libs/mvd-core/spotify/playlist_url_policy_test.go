package spotify

import "testing"

func TestPlaylistID(t *testing.T) {
	const id = "37i9dQZF1DXcBWIGoYBM5M"
	good := []string{
		"https://open.spotify.com/playlist/" + id,
		"https://open.spotify.com/playlist/" + id + "?si=abc123",
		"https://open.spotify.com/intl-pt/playlist/" + id + "/",
		"https://open.spotify.com/embed/playlist/" + id + "?utm=x",
		"  spotify:playlist:" + id + " ",
	}
	for _, link := range good {
		if got, ok := PlaylistID(link); !ok || got != id {
			t.Errorf("PlaylistID(%q) = %q, %v; want %q", link, got, ok, id)
		}
	}

	for _, link := range []string{
		"",
		"https://open.spotify.com/track/" + id,
		"https://open.spotify.com/album/" + id,
		"https://example.com/playlist/" + id,
		"https://www.youtube.com/playlist?list=PL123",
		"spotify:playlist:",
		"spotify:track:" + id,
	} {
		if got, ok := PlaylistID(link); ok {
			t.Errorf("PlaylistID(%q) = %q, want no match", link, got)
		}
	}
}
