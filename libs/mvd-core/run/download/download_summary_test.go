package download

import (
	"testing"

	"youtube-downloader/libs/mvd-core/runstate"
)

func TestTheDownloadingSummaryGivesTheSongsAndTheSpeedAndNeverTheParts(t *testing.T) {
	got := downloadingSummary(runstate.Tally{Downloading: 4, Parts: 9, Speed: 12.4 * 1024 * 1024})
	if want := "Downloading 4 songs · 12.4MiB/s"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := downloadingSummary(runstate.Tally{Downloading: 1, Parts: 1, Speed: 0}); got != "Downloading 1 song" {
		t.Errorf("one song with no known speed: %q", got)
	}
}
