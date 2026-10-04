package snapshot

import (
	"encoding/json"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func sampleState() *runstate.State {
	s := runstate.New([]engine.PlaylistSource{{Index: 0, URL: "https://a"}, {Index: 1, URL: "https://b"}})
	s.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "v0", Title: "zero", Channel: "Band"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "v1", Title: "one", Channel: "Band"},
	}})
	s.Apply(engine.EvPlaylistFailed{Playlist: 1, Err: "boom"})
	s.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone, TargetID: "official0", Official: true})
	s.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDownloading, TargetID: "v1"})
	s.Apply(engine.EvProgress{Entry: 1, Percent: 40, Downloaded: 400, Total: 1000, Speed: 100, ETA: 6})
	return s
}

func TestFromMapsThePlaylistsAndEntries(t *testing.T) {
	snap := From(sampleState(), 7)

	if snap.Version != 7 {
		t.Errorf("version = %d, want 7", snap.Version)
	}
	if len(snap.Playlists) != 2 || len(snap.Entries) != 2 {
		t.Fatalf("want 2 playlists and 2 entries, got %d and %d", len(snap.Playlists), len(snap.Entries))
	}

	a := snap.Playlists[0]
	if a.Title != "A" || !a.Listed || a.Total != 2 || a.Finished != 1 || a.Active != 1 || len(a.Entries) != 2 {
		t.Errorf("playlist A wrong: %+v", a)
	}
	if b := snap.Playlists[1]; b.Err != "boom" || !b.Listed || b.Total != 0 {
		t.Errorf("playlist B wrong: %+v", b)
	}

	done := snap.Entries[0]
	if done.State != "done" || !done.Official || done.VideoID != "v0" || done.TargetID != "official0" || done.Percent != 100 {
		t.Errorf("done entry wrong: %+v", done)
	}
	running := snap.Entries[1]
	if running.State != "downloading" || running.Percent != 40 || running.Downloaded != 400 || running.TotalBytes != 1000 || running.Speed != 100 || running.ETA != 6 {
		t.Errorf("running entry wrong: %+v", running)
	}

	if tl := snap.Tally; tl.Total != 2 || tl.Done != 1 || tl.Running != 1 || tl.Official != 1 {
		t.Errorf("tally wrong: %+v", tl)
	}
}

func TestFromSkipsTheGapWhereAnEntryHasNotArrivedYet(t *testing.T) {
	s := runstate.New([]engine.PlaylistSource{{Index: 0, URL: "https://a"}})
	// Id 2 arrives before ids 0 and 1: the state leaves holes until they do.
	s.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{{ID: 2, Playlist: 0, Index: 1, VideoID: "v2"}}})

	snap := From(s, 1)

	if len(snap.Entries) != 1 || snap.Entries[0].ID != 2 {
		t.Errorf("want only entry 2, got %+v", snap.Entries)
	}
}

func TestFromEncodesEmptyListsAsArraysNotNull(t *testing.T) {
	raw, err := json.Marshal(From(runstate.New(nil), 0))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"playlists", "entries"} {
		if string(decoded[key]) != "[]" {
			t.Errorf("%s encoded as %s, want []", key, decoded[key])
		}
	}
}

func TestFromEncodesTheFieldNamesTheBrowserDependsOn(t *testing.T) {
	raw, err := json.Marshal(From(sampleState(), 3))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Playlists []map[string]any `json:"playlists"`
		Entries   []map[string]any `json:"entries"`
		Tally     map[string]any   `json:"tally"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"index", "url", "title", "err", "listed", "total", "finished", "failed", "active", "entries"} {
		if _, ok := decoded.Playlists[0][key]; !ok {
			t.Errorf("playlist has no %q key: %v", key, decoded.Playlists[0])
		}
	}
	for _, key := range []string{"id", "playlist", "index", "videoId", "targetId", "title", "channel", "state", "official", "err", "percent", "downloaded", "totalBytes", "speed", "eta"} {
		if _, ok := decoded.Entries[0][key]; !ok {
			t.Errorf("entry has no %q key: %v", key, decoded.Entries[0])
		}
	}
	for _, key := range []string{"total", "queued", "running", "done", "official", "duplicate", "failed", "retried"} {
		if _, ok := decoded.Tally[key]; !ok {
			t.Errorf("tally has no %q key: %v", key, decoded.Tally)
		}
	}
}

func TestFromDoesNotShareTheStatesSlices(t *testing.T) {
	s := sampleState()
	snap := From(s, 1)

	snap.Playlists[0].Entries[0] = 99

	if s.Playlists[0].Entries[0] == 99 {
		t.Error("changing the snapshot changed the live state")
	}
}
