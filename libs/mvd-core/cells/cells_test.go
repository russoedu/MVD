package cells

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDisplay(t *testing.T) {
	cases := map[string]string{
		"Música Eletrônica 1990s 🔊 1999, 1998": "Música Eletrônica 1990s  1999, 1998",
		"Música": "Música", // decomposed accent becomes one code point
		"Stardust 日本語タイトル 한국어":                         "Stardust 日本語タイトル 한국어",
		"tab\there\x00\x1b[31mred":                     "tab here[31mred",
		"flags 🇧🇷 and skin 👍🏽 and zwj 👨‍👩‍👧":           "flags  and skin  and zwj",
		"keep ♪ ★ ☆ ✓ symbols":                         "keep ♪ ★ ☆ ✓ symbols",
		"heavy ✅ ❌ ⭐ gone":                             "heavy    gone",
		"variation ▶️ selector":                        "variation ▶ selector",
		"":                                             "",
		"  padded  ":                                   "padded",
		"[download] Destination: ../DJ/new/Song 🎵.mp4": "[download] Destination: ../DJ/new/Song .mp4",
	}
	for in, want := range cases {
		if got := Display(in); got != want {
			t.Errorf("Display(%q)\n want %q\n got  %q", in, want, got)
		}
	}
	out := Display("mix 🔊 of 日本 and 🎵 emoji ✅")
	if w := ansi.StringWidth(out); w != len([]rune(out))+2 {
		t.Errorf("unexpected width %d for %q", w, out)
	}
}

func TestTruncateEndsInAnEllipsis(t *testing.T) {
	if got := Truncate("Stardust", 5); got != "Star…" {
		t.Errorf("got %q", got)
	}
	if got := Truncate("Stardust", 20); got != "Stardust" {
		t.Errorf("a line that fits is kept, got %q", got)
	}
}

func TestFitPadsAndCuts(t *testing.T) {
	for _, w := range []int{3, 8, 12} {
		if got := ansi.StringWidth(Fit("Stardust", w)); got != w {
			t.Errorf("Fit to %d is %d cells wide", w, got)
		}
	}
}
