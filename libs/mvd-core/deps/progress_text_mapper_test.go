package deps

import (
	"errors"
	"strings"
	"testing"
)

func TestEachKindOfStepHasItsOwnLineOnTheTerminal(t *testing.T) {
	cases := []struct {
		event Event
		want  string
	}{
		{Event{Kind: EventMissing, Names: []string{"yt-dlp", "ffmpeg"}, Dir: "bin"}, "Missing dependency/dependencies detected: yt-dlp, ffmpeg"},
		{Event{Kind: EventDownloading, Name: "yt-dlp", URL: "https://example.test/yt-dlp"}, "Downloading yt-dlp from https://example.test/yt-dlp..."},
		{Event{Kind: EventInstalled, Name: "yt-dlp"}, "[OK] Successfully installed yt-dlp!"},
		{Event{Kind: EventFailed, Name: "yt-dlp", Err: errors.New("no network")}, "[!] Could not install yt-dlp: no network"},
		{Event{Kind: EventUpToDate, Name: "yt-dlp", Detail: "2026.08.19"}, "yt-dlp is up to date (2026.08.19)"},
		{Event{Kind: EventUpdated, Name: "yt-dlp", Detail: "2026.09.01"}, "[OK] Updated yt-dlp to 2026.09.01"},
		{Event{Kind: EventUpdateFailed, Name: "yt-dlp", Err: errors.New("rate limited")}, "[!] Could not update yt-dlp: rate limited"},
	}

	for _, c := range cases {
		if got := progressText(c.event); !strings.Contains(got, c.want) {
			t.Errorf("kind %d: %q does not contain %q", c.event.Kind, got, c.want)
		}
	}
}

func TestAnUnknownKindOfStepShowsNothing(t *testing.T) {
	if got := progressText(Event{Kind: EventKind(99)}); got != "" {
		t.Errorf("got %q, want nothing", got)
	}
}
