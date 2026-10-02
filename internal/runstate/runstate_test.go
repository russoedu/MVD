package runstate

import (
	"testing"

	"youtube-downloader/internal/engine"
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
