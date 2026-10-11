package list

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
		name string
		urls []string
		keys []string
	}{
		{"list-empty", nil, nil},
		{"list-three", []string{"https://youtube.com/playlist?list=A", "https://youtube.com/playlist?list=B", `C:\songs.txt`}, nil},
		{"list-confirm-quit", []string{"https://youtube.com/playlist?list=A"}, []string{"esc"}},
	}
	for _, c := range cases {
		look.Apply(config.DefaultThemeColors())
		m := New(c.urls).SetSize(w, h)
		for _, msg := range keys.Sequence(c.keys...) {
			m, _, _ = m.Update(msg)
		}
		drawing.Check(t, "testdata", c.name, w, m.View(w, h))
	}
}
