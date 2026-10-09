package runstate

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestApplyAndTally(t *testing.T) {
	s := New([]engine.PlaylistSource{{Index: 0, URL: "https://a"}, {Index: 1, URL: "https://b"}})
	s.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "v0", Title: "zero"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "v1", Title: "one"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "v2", Title: "two"},
	}})
	s.Apply(engine.EvPlaylistFailed{Playlist: 1, Err: "boom"})
	s.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone, TargetID: "off", Official: true})
	s.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDownloading, TargetID: "v1"})
	s.Apply(engine.EvProgress{Entry: 1, Percent: 40, Total: 100, Downloaded: 40, Speed: 10, ETA: 6})
	s.Apply(engine.EvLog{Playlist: 0, Entry: 1, Line: "hello"})
	s.Apply(engine.EvLog{Playlist: 0, Entry: -1, Line: "listed"})
	if id := s.Apply(engine.EvEntryState{Entry: 99, State: engine.StateDone}); id != -1 {
		t.Error("unknown entry should be ignored")
	}

	tl := s.Tally()
	if tl.Total != 3 || tl.Done != 1 || tl.Official != 1 || tl.Running != 1 || tl.Queued != 1 {
		t.Errorf("unexpected tally %+v", tl)
	}
	if s.Playlists[0].Title != "A" || s.Playlists[1].Err != "boom" || !s.Playlists[1].Listed {
		t.Errorf("playlists not updated: %+v %+v", s.Playlists[0], s.Playlists[1])
	}
	if en := s.Entry(1); en.Percent != 40 || len(en.Log) != 1 {
		t.Errorf("entry 1 not updated: %+v", en)
	}
	if got := s.Playlists[0].Log; len(got) != 2 || got[0] != "[02] hello" || got[1] != "listed" {
		t.Errorf("playlist log wrong: %v", got)
	}
	finished, failed, active := s.PlaylistTally(s.Playlists[0])
	if finished != 1 || failed != 0 || active != 1 {
		t.Errorf("playlist tally wrong: %d %d %d", finished, failed, active)
	}
	s.Apply(engine.EvIdle{})
	if !s.Idle {
		t.Error("idle not applied")
	}

	// A failed -> queued transition counts as a retry.
	s.Apply(engine.EvEntryState{Entry: 2, State: engine.StateFailed, Err: "boom"})
	s.Apply(engine.EvEntryState{Entry: 2, State: engine.StateQueued})
	s.Apply(engine.EvEntryState{Entry: 2, State: engine.StateDone})
	if got := s.Tally().Retried; got != 1 {
		t.Errorf("retried = %d, want 1", got)
	}
}

func TestApplyPlaylistAddedWhileRunning(t *testing.T) {
	s := New([]engine.PlaylistSource{{Index: 0, URL: "https://a"}})
	s.Apply(engine.EvIdle{})
	if !s.Idle {
		t.Fatal("setup: the run should be idle")
	}

	s.Apply(engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: 1, URL: "https://b"}})

	if len(s.Playlists) != 2 || s.Playlists[1].Index != 1 || s.Playlists[1].URL != "https://b" || s.Playlists[1].Title != "https://b" {
		t.Fatalf("playlist not added as the next index: %+v", s.Playlists)
	}
	if s.Idle {
		t.Error("a run that has work again is not idle")
	}

	// The same event again is already folded in: it must not duplicate the playlist.
	s.Apply(engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: 1, URL: "https://b"}})
	if len(s.Playlists) != 2 {
		t.Errorf("a repeated event added a playlist: %d", len(s.Playlists))
	}

	// The engine's later events about it land on it like on any other playlist.
	s.Apply(engine.EvPlaylistListed{Playlist: 1, Title: "B", Entries: []engine.EntryInfo{{ID: 0, Playlist: 1, Index: 1, VideoID: "v0", Title: "zero"}}})
	if pl := s.Playlists[1]; pl.Title != "B" || !pl.Listed || len(pl.Entries) != 1 {
		t.Errorf("listing did not reach the added playlist: %+v", pl)
	}
	s.Apply(engine.EvLog{Playlist: 1, Entry: -1, Line: "listed"})
	if got := s.Playlists[1].Log; len(got) != 1 || got[0] != "listed" {
		t.Errorf("log did not reach the added playlist: %v", got)
	}
}

func TestHumanFormats(t *testing.T) {
	cases := map[int64]string{500: "500B", 2048: "2.0KiB", 117833728: "112.4MiB", 3 << 30: "3.0GiB"}
	for in, want := range cases {
		if got := HumanBytes(in); got != want {
			t.Errorf("HumanBytes(%d) = %q, want %q", in, got, want)
		}
	}
	if HumanETA(9) != "0:09" || HumanETA(3725) != "1:02:05" || HumanETA(-1) != "--" {
		t.Error("HumanETA wrong")
	}
	if HumanSpeed(0) != "--" || HumanSpeed(8.5*1024*1024) != "8.5MiB/s" {
		t.Error("HumanSpeed wrong")
	}
	if got := ProgressLine(&Entry{Percent: 34.2, Total: 117833728, Speed: 8.1 * 1024 * 1024, ETA: 9}); got != " 34.2% of 112.4MiB at 8.1MiB/s ETA 0:09" {
		t.Errorf("ProgressLine wrong: %q", got)
	}
}

func TestTheTallyAddsUpWhatIsComingInNow(t *testing.T) {
	s := New(nil)
	s.Entries = []*Entry{
		{State: engine.StateDownloading, Speed: 2_000_000, Parts: 4},
		{State: engine.StateDownloading, Speed: 1_000_000},
		{State: engine.StateMerging, Speed: 500_000, Parts: 2},
		{State: engine.StateQueued},
	}

	tally := s.Tally()
	if tally.Downloading != 2 || tally.Parts != 5 || tally.Speed != 3_000_000 {
		t.Errorf("downloading %d, parts %d, speed %v: a merging file is not coming in", tally.Downloading, tally.Parts, tally.Speed)
	}
}
