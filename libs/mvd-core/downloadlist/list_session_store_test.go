package downloadlist

import (
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

const listURL = "https://youtube.com/playlist?list=A"

func snap(id string, index int, state engine.EntryState, target string) engine.EntrySnapshot {
	return engine.EntrySnapshot{
		PlannedEntry: engine.PlannedEntry{
			Playlist:    "A",
			PlaylistURL: listURL,
			Upload:      ytdlp.PlaylistEntry{ID: id, Title: "Song " + id, PlaylistIndex: index},
			TargetID:    target,
			Kind:        engine.KindOriginal,
		},
		State: state,
	}
}

func TestANewListHasItsAddressesAndNoFileYet(t *testing.T) {
	s := New([]string{"a", "b", "a", ""})
	if got := s.Sources(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("sources %v", got)
	}
	if s.Path() != "" {
		t.Errorf("a new list has no file, got %q", s.Path())
	}
	if err := s.Save(""); err == nil {
		t.Error("saving a list that has no file must ask for one")
	}
}

func TestEverythingIsToBeListedBeforeTheRunHasSeenIt(t *testing.T) {
	sources, plan := New([]string{listURL}).Inputs()
	if len(sources) != 1 || len(plan) != 0 {
		t.Fatalf("sources %v, plan %v", sources, plan)
	}
}

func TestADownloadedSongIsNotGivenAgainButTheOthersAre(t *testing.T) {
	s := New([]string{listURL})
	s.Apply([]engine.EntrySnapshot{
		snap("one", 1, engine.StateDone, "one"),
		snap("two", 2, engine.StateDownloading, "two"),
		snap("three", 3, engine.StateFailed, "three"),
		snap("four", 4, engine.StateQueued, ""), // not looked up yet: nothing to download
	})
	sources, plan := s.Inputs()
	if len(sources) != 0 {
		t.Errorf("a listed item must not be listed again: %v", sources)
	}
	if len(plan) != 2 || plan[0].Upload.ID != "two" || plan[1].Upload.ID != "three" {
		t.Fatalf("plan %+v, want the unfinished songs that have a video", plan)
	}
	song := s.File().Songs()
	if !song[0].Downloaded || song[0].State != playlistfile.StateDone || song[2].State != playlistfile.StateFailed {
		t.Errorf("states %+v", song)
	}
}

func TestWhatThePersonDecidedSurvivesTheRunReportingAgain(t *testing.T) {
	s := New([]string{listURL})
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StatePlanned, "one")})

	f := s.File()
	f.Items[0].Songs[0].Decision = playlistfile.DecisionSkipped
	s.file = f

	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StatePlanned, "one")})
	if got := s.File().Songs()[0].Decision; got != playlistfile.DecisionSkipped {
		t.Errorf("decision %q was lost", got)
	}
	if _, plan := s.Inputs(); len(plan) != 0 {
		t.Errorf("a skipped song must not be given to a run: %+v", plan)
	}
}

func TestDownloadedStaysDownloadedWhateverTheRunSaysLater(t *testing.T) {
	s := New([]string{listURL})
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StateDone, "one")})
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StateQueued, "one")})
	if !s.File().Songs()[0].Downloaded {
		t.Error("a song that was downloaded must stay downloaded")
	}
}

func TestACancelledSongIsWaitingNotFailed(t *testing.T) {
	s := New([]string{listURL})
	cancelled := snap("one", 1, engine.StateFailed, "one")
	cancelled.Err = "cancelled"
	s.Apply([]engine.EntrySnapshot{cancelled})
	if got := s.File().Songs()[0]; got.State != playlistfile.StateWaiting || got.Error != "" {
		t.Errorf("a song cancelled when the run was closed is not a failure: %+v", got)
	}
}

func TestALinkAddedWhileRunningBecomesAnItem(t *testing.T) {
	s := New(nil)
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StateQueued, "one")})
	if got := s.Sources(); len(got) != 1 || got[0] != listURL {
		t.Errorf("sources %v", got)
	}
}

func TestSavingAndOpeningGiveTheListBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mine.mvd")
	s := New([]string{listURL})
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StateDone, "one"), snap("two", 2, engine.StateQueued, "two")})
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	if s.Path() != path {
		t.Errorf("path %q", s.Path())
	}

	again, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, plan := again.Inputs(); len(plan) != 1 || plan[0].Upload.ID != "two" {
		t.Errorf("plan %+v: only the unfinished song carries on", plan)
	}
}

func TestAnUnchangedListIsNotWrittenAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mine.mvd")
	s := New([]string{listURL})
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	before, _ := filepath.Glob(path)
	if err := s.flush(); err != nil || len(before) != 1 {
		t.Fatal(err)
	}
	if string(encode(s.file)) != string(s.saved) {
		t.Error("flush of an unchanged list should find nothing to write")
	}
}

func TestSetSourcesKeepsTheSongsOfWhatStays(t *testing.T) {
	s := New([]string{listURL, "other"})
	s.Apply([]engine.EntrySnapshot{snap("one", 1, engine.StateDone, "one")})
	s.SetSources([]string{"new", listURL})
	got := s.File()
	if len(got.Items) != 2 || got.Items[0].Source != "new" || len(got.Items[1].Songs) != 1 {
		t.Errorf("items %+v", got.Items)
	}
}

func TestThePreviewShowsEachSongInTheStateTheFileKept(t *testing.T) {
	s := New([]string{listURL, "https://not.looked.at"})
	s.Apply([]engine.EntrySnapshot{
		snap("one", 1, engine.StateDone, "one"),
		snap("two", 2, engine.StateFailed, "two"),
		snap("three", 3, engine.StatePlanned, "three"),
	})
	state := s.Preview()
	if len(state.Playlists) != 2 || len(state.Playlists[0].Entries) != 3 || len(state.Playlists[1].Entries) != 0 {
		t.Fatalf("playlists %+v", state.Playlists)
	}
	tally := state.Tally()
	if tally.Done != 1 || tally.Failed != 1 || tally.Planned != 1 {
		t.Errorf("tally %+v", tally)
	}
	if state.Playlists[0].Title != "A" || state.Playlists[1].Title != "https://not.looked.at" {
		t.Errorf("titles %q %q", state.Playlists[0].Title, state.Playlists[1].Title)
	}
}
