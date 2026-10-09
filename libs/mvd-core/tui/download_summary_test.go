package tui

import (
	"testing"

	"youtube-downloader/libs/mvd-core/runstate"
)

func TestTheDownloadingSummaryGivesSpeedFilesAndParts(t *testing.T) {
	got := downloadingSummary(runstate.Tally{Downloading: 4, Parts: 9, Speed: 12.4 * 1024 * 1024})
	want := "↓ 12.4MiB/s · 4 files · 9 parts"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := downloadingSummary(runstate.Tally{Downloading: 1, Parts: 1, Speed: 0}); got != "↓ -- · 1 file · 1 part" {
		t.Errorf("one file with no known speed: %q", got)
	}
}
