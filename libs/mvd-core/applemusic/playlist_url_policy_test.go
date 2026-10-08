package applemusic

import "testing"

func TestPlaylistURL(t *testing.T) {
	const want = "https://music.apple.com/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb"
	for _, link := range []string{
		want,
		want + "?l=pt",
		want + "#section",
		"http://music.apple.com/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb",
		"  " + want + "  ",
	} {
		if got, ok := PlaylistURL(link); !ok || got != want {
			t.Errorf("PlaylistURL(%q) = %q, %v; want %q", link, got, ok, want)
		}
	}

	if got, ok := PlaylistURL("https://music.apple.com/pt/playlist/pl.f4d106fed2bd41149aaacabb233eb5eb"); !ok || got != "https://music.apple.com/pt/playlist/pl.f4d106fed2bd41149aaacabb233eb5eb" {
		t.Errorf("a link without the name part should work, got %q, %v", got, ok)
	}

	for _, link := range []string{
		"",
		"https://music.apple.com/us/album/some-album/1234567890",
		"https://music.apple.com/us/playlist/todays-hits/notanid",
		"https://example.com/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb",
		"https://itunes.apple.com/us/playlist/todays-hits/pl.f4d106fed2bd41149aaacabb233eb5eb",
		"https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
	} {
		if got, ok := PlaylistURL(link); ok {
			t.Errorf("PlaylistURL(%q) = %q, want no match", link, got)
		}
	}
}
