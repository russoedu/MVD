package colours

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/drawing"
	"youtube-downloader/libs/mvd-core/testing/keys"
)

// TestTheScreenKeepsItsLook compares the screen in four states with the reference
// drawings of the hand written screen it replaced.
func TestTheScreenKeepsItsLook(t *testing.T) {
	const w, h = 100, 30
	cases := []struct {
		name string
		keys []string
	}{
		{"colours-first", nil},
		{"colours-fourth", []string{"down", "down", "down"}},
		{"colours-editing", []string{"down", "enter"}},
		{"colours-error", []string{"enter", "backspace", "z", "enter"}},
	}
	for _, c := range cases {
		look.Apply(config.DefaultThemeColors())
		m := New(config.Default(`C:\Music`, `C:\Music\logs`)).SetSize(w, h)
		for _, msg := range keys.Sequence(c.keys...) {
			m, _, _, _ = m.Update(msg)
		}
		drawing.Check(t, "testdata", c.name, w, m.View(w, h))
	}
}
