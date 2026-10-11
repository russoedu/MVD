package songs

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/drawing"
)

// TestTheScreenKeepsItsLook compares the screen in two states with the reference
// drawings of the hand written screen it replaced.
func TestTheScreenKeepsItsLook(t *testing.T) {
	const w, h = 100, 30
	cases := []struct{ name, text string }{
		{"songs-empty", ""},
		{"songs-two", "Daft Punk - One More Time\nbad line\nAlexia - Uh La La La"},
	}
	for _, c := range cases {
		look.Apply(config.DefaultThemeColors())
		m := New(`C:\Music\songs`).SetSize(w, h).WithText(c.text)
		drawing.Check(t, "testdata", c.name, w, m.View(w, h))
	}
}
