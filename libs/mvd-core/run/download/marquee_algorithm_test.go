package download

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestATextThatFitsStaysStill(t *testing.T) {
	if got := marquee("short", 10, 500); got != "short" {
		t.Errorf("got %q", got)
	}
}

func TestALongTextScrollsAfterAPauseAndWrapsRound(t *testing.T) {
	text := "abcdefghijklmnopqrstuvwxyz"
	if got := marquee(text, 8, 0); got != "abcdefgh" {
		t.Errorf("at the start %q", got)
	}
	later := marquee(text, 8, marqueeEvery*(marqueePause+3))
	if later != "defghijk" {
		t.Errorf("after three moves %q", later)
	}
	for tick := 0; tick < 2000; tick++ {
		if w := ansi.StringWidth(marquee(text, 8, tick)); w != 8 {
			t.Fatalf("tick %d is %d wide", tick, w)
		}
	}
}

func TestPlaceholdersWrapToTheColumn(t *testing.T) {
	for _, line := range wrapText("Add music: playlists or videos, with the Add link button", 12) {
		if ansi.StringWidth(line) > 12 || strings.TrimSpace(line) == "" {
			t.Errorf("line %q", line)
		}
	}
}
