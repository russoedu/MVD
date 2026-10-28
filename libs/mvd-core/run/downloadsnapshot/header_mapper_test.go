package downloadsnapshot

import (
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/musicsearch"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
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

func TestASearchByNameIsCalledWhatItAsksFor(t *testing.T) {
	address := musicsearch.Query{Artist: "Daft Punk", Album: "Discovery"}.Address()
	if got := shortAddress(address); got != "Daft Punk – Discovery" {
		t.Errorf("a search is called %q", got)
	}
	if got := shortAddress("https://www.youtube.com/watch?v=x"); got != "youtube.com/watch?v=x" {
		t.Errorf("a web address is called %q", got)
	}
}

func TestAPauseThatCameAfterTheLastSongIsAFinishedRun(t *testing.T) {
	state := runstate.New(nil)
	state.Idle = true
	s := downloadsession.Session{State: state, Paused: true}
	if got := stateLabel(s, runstate.Tally{}); !strings.HasPrefix(got, "FINISHED") {
		t.Errorf("a held run with nothing left says %q", got)
	}
	s.State.Idle = false
	if got := stateLabel(s, runstate.Tally{}); !strings.HasPrefix(got, "PAUSED") {
		t.Errorf("a held run with work left says %q", got)
	}
}
