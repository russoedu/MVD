package folder

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/drawing"
)

// TestTheScreenKeepsItsLook compares the screen in three states with the reference
// drawings of the hand written screen it replaced.
func TestTheScreenKeepsItsLook(t *testing.T) {
	const w, h = 100, 30
	cases := []struct {
		name     string
		creating bool
		err      string
	}{
		{"folder-browse", false, ""},
		{"folder-creating", true, ""},
		{"folder-error", false, `open C:\Users\me\Music: access denied`},
	}
	for _, c := range cases {
		look.Apply(config.DefaultThemeColors())
		m := New(".").Listing(`C:\Users\me\Music`, []string{"90s", "Dance", "Rock", "Soundtracks"}, 1, c.err, c.creating)
		drawing.Check(t, "testdata", c.name, w, m.SetSize(w, h).View(w, h))
	}
}
