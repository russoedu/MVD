package runstate

import (
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
)

func oneSongState() *State {
	s := New([]engine.PlaylistSource{{Index: 0, URL: "https://a"}})
	s.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "a", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, VideoID: "v", Title: "song"}}})

	return s
}

func TestTheTimeOfAFinishedRunStopsWhereItWentIdle(t *testing.T) {
	s := oneSongState()
	s.Started = time.Now().Add(-10 * time.Second)
	s.Apply(engine.EvIdle{})
	first := s.Elapsed()
	if first < 9*time.Second || first > 12*time.Second {
		t.Fatalf("elapsed at idle is %v, want about 10s", first)
	}
	time.Sleep(60 * time.Millisecond)
	if again := s.Elapsed(); again != first {
		t.Errorf("a finished run's time moved from %v to %v", first, again)
	}
}

func TestTheClockRunsAgainWhenSongsAreTriedAgain(t *testing.T) {
	s := oneSongState()
	s.Started = time.Now().Add(-10 * time.Second)
	s.Apply(engine.EvIdle{})
	frozen := s.Elapsed()
	s.Apply(engine.EvEntryState{Entry: 0, State: engine.StateQueued})
	if s.Idle {
		t.Fatal("a song queued again should end the idle state")
	}
	time.Sleep(60 * time.Millisecond)
	if s.Elapsed() <= frozen {
		t.Error("the clock did not start again")
	}
}
