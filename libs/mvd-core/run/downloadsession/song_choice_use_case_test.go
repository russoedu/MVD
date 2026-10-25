package downloadsession

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

// fakeRun is a run that holds its songs for choosing and records what it was told.
type fakeRun struct {
	events   chan interface{}
	released []int
	skipped  []int
	swapped  map[int]string
	retried  []int
	added    []string
}

func newFakeRun() *fakeRun {
	return &fakeRun{events: make(chan interface{}), swapped: map[int]string{}}
}

func (f *fakeRun) Events() <-chan interface{} { return f.events }
func (f *fakeRun) Sources() []engine.PlaylistSource {
	return []engine.PlaylistSource{{Index: 0, URL: "https://a"}}
}
func (f *fakeRun) LogPath() string                { return "" }
func (f *fakeRun) Retry(id int) bool              { f.retried = append(f.retried, id); return true }
func (f *fakeRun) RetryPlaylist(int) int          { return 0 }
func (f *fakeRun) Release(id int, _ string) bool  { f.released = append(f.released, id); return true }
func (f *fakeRun) ReleaseAll() int                { return 2 }
func (f *fakeRun) Skip(id int) bool               { f.skipped = append(f.skipped, id); return true }
func (f *fakeRun) Swap(id int, video string) bool { f.swapped[id] = video; return true }
func (f *fakeRun) AddSource(u string) (engine.PlaylistSource, bool) {
	f.added = append(f.added, u)

	return engine.PlaylistSource{}, true
}

// sessionWithSongs is a session on a run that listed three songs waiting to be chosen.
func sessionWithSongs(run *fakeRun) Session {
	s := New(run)
	entries := []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "a", Title: "Alpha"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "b", Title: "Bravo"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "c", Title: "Charlie"},
	}
	s.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "list", Entries: entries})
	for _, e := range entries {
		s.Apply(engine.EvEntryState{Entry: e.ID, State: engine.StatePlanned})
	}
	s.Pane = PaneEntries

	return s
}

func TestAcceptingWorksOnTheSongsTickedElseOnTheOneTheCursorIsOn(t *testing.T) {
	run := newFakeRun()
	s := sessionWithSongs(run)

	s.Move(1)
	s.Accept()
	if len(run.released) != 1 || run.released[0] != 1 {
		t.Fatalf("with nothing ticked the cursor's song is accepted, got %v", run.released)
	}

	run.released = nil
	s.TickSong(0, false)
	s.TickSong(2, false)
	s.Accept()
	if len(run.released) != 2 || run.released[0] != 0 || run.released[1] != 2 {
		t.Errorf("the ticked songs are accepted, got %v", run.released)
	}
	if len(s.Picked) != 0 {
		t.Error("the ticks should go once the songs are accepted")
	}
}

func TestATickWithExtendTicksEverySongFromTheLastTickToIt(t *testing.T) {
	s := sessionWithSongs(newFakeRun())
	s.TickSong(0, false)
	s.TickSong(2, true)
	if len(s.Picked) != 3 {
		t.Errorf("a range from the first to the last song ticks all three, got %v", s.Picked)
	}
	if s.PickState() != 2 {
		t.Errorf("all songs are ticked, the state says %d", s.PickState())
	}
}

func TestRejectingAndSwappingSayWhyWhenTheyCannotBeDone(t *testing.T) {
	run := newFakeRun()
	s := sessionWithSongs(run)

	s.Reject()
	if len(run.skipped) != 1 {
		t.Fatalf("the song under the cursor should be rejected, got %v", run.skipped)
	}

	s.TickSong(0, false)
	s.TickSong(1, false)
	s.BeginSwap()
	if s.Swapping || s.Notice != "swap the video of one song at a time" {
		t.Errorf("two songs cannot be swapped at once: swapping %v, notice %q", s.Swapping, s.Notice)
	}

	s.Picked = nil
	s.BeginSwap()
	s.SetSwapText("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	s.CommitSwap()
	if run.swapped[0] != "dQw4w9WgXcQ" || s.Swapping {
		t.Errorf("the swap should have reached the run: %v, swapping %v", run.swapped, s.Swapping)
	}
}

func TestOpeningAVideoIsAnEffectForTheHost(t *testing.T) {
	s := sessionWithSongs(newFakeRun())
	eff := s.OpenVideo()
	open, ok := eff.(OpenURL)
	if !ok || open.URL != "https://www.youtube.com/watch?v=a" {
		t.Errorf("opening the first song gave %#v", eff)
	}

	s.TickSong(0, false)
	s.TickSong(1, false)
	if eff := s.OpenVideo(); eff != nil || s.Notice != "open one video at a time" {
		t.Errorf("two songs cannot be opened at once: %#v, %q", eff, s.Notice)
	}
}

func TestAskingToQuitWhileTheDownloadsGoAsksFirst(t *testing.T) {
	s := sessionWithSongs(newFakeRun())
	if eff := s.RequestQuit(); eff != nil || !s.ConfirmQuit {
		t.Fatalf("a run that goes on should ask first: %#v, confirm %v", eff, s.ConfirmQuit)
	}
	if eff := s.AnswerQuit(false); eff != nil || s.ConfirmQuit {
		t.Errorf("answering no should stay: %#v, confirm %v", eff, s.ConfirmQuit)
	}
	s.RequestQuit()
	if eff := s.AnswerQuit(true); eff != (Quit{}) {
		t.Errorf("answering yes should leave: %#v", eff)
	}

	s.State.Idle = true
	if eff := s.RequestQuit(); eff != (Quit{}) {
		t.Errorf("a finished run leaves at once: %#v", eff)
	}
}

func TestAHostedScreenLeavesQuittingAndAddingToTheWorkspace(t *testing.T) {
	s := NewHosted()
	if eff := s.RequestQuit(); eff != nil || s.ConfirmQuit {
		t.Errorf("a hosted screen does not quit by itself: %#v", eff)
	}
	s.BeginAdding()
	if s.Adding || s.Notice != "" {
		t.Error("a hosted screen does not open the box for links")
	}
}

func TestAddingLinksQueuesEveryLineAndFollowsTheRun(t *testing.T) {
	run := newFakeRun()
	s := sessionWithSongs(run)
	s.Follow = false
	s.BeginAdding()
	if !s.Adding {
		t.Fatal("the box should open on a run that takes links")
	}
	s.SetLinksText("https://one\n\n  https://two  \n")
	s.CommitLinks()
	if len(run.added) != 2 || run.added[1] != "https://two" || s.Adding || !s.Follow {
		t.Errorf("added %v, adding %v, follow %v", run.added, s.Adding, s.Follow)
	}
	if s.Notice != "added 2 links" {
		t.Errorf("notice %q", s.Notice)
	}
}

func TestTheOrderOfTheColumnsTheyArrangeKeepsTheOnesNotMentioned(t *testing.T) {
	s := Session{}
	s.MoveColumn(0, 2)
	got := s.ArrangedColumns()
	if len(got) != len(Configurable) || got[2] != ColumnArtist {
		t.Errorf("moving the first column to the third gave %v", got)
	}
	s.ToggleColumn(ColumnAlbum)
	if !s.HiddenCols[ColumnAlbum] {
		t.Error("toggling a column should hide it")
	}
	s.ToggleColumn(ColumnAlbum)
	if s.HiddenCols[ColumnAlbum] {
		t.Error("toggling it again should show it")
	}
}
