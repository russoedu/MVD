package advanced

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/drawing"
	"youtube-downloader/libs/mvd-core/testing/keys"
)

// TestTheScreenKeepsItsLook compares the screen in three states with the reference
// drawings of the hand written screen it replaced.
func TestTheScreenKeepsItsLook(t *testing.T) {
	const w, h = 100, 30
	cases := []struct {
		name  string
		extra []string
		keys  []string
	}{
		{"advanced-first", nil, nil},
		{"advanced-args", []string{"--proxy", "http://x"}, []string{"down"}},
		{"advanced-editing", nil, []string{"enter"}},
	}
	for _, c := range cases {
		look.Apply(config.DefaultThemeColors())
		cfg := config.Default(`C:\Music`, `C:\Music\logs`)
		cfg.ExtraArgs = c.extra
		m := New(cfg).SetSize(w, h)
		for _, msg := range keys.Sequence(c.keys...) {
			m, _, _, _ = m.Update(msg)
		}
		drawing.Check(t, "testdata", c.name, w, m.View(w, h))
	}
}
