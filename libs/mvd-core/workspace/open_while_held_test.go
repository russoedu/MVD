package workspace

import (
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// otherList saves a list with one link and gives its path.
func otherList(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "other.mvd")
	if err := downloadlist.New([]string{"https://youtu.be/other"}).Save(path); err != nil {
		t.Fatal(err)
	}

	return path
}

// download makes the run report one song that is coming in now.
func download(m Model) Model {
	state := runstate.New([]engine.PlaylistSource{{Index: 0, URL: "https://youtu.be/x"}})
	m.dl.Show(state)
	state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "list", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, VideoID: "a", Title: "A"}}})
	state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDownloading})

	return m
}

func TestAListOpensWhileTheRunOnlyHoldsTheSongsForChoosing(t *testing.T) {
	m, fake := started(t)
	path := otherList(t)

	next, cmd := m.Update(OpenListMsg{Path: path})
	m = next.(Model)
	if !m.life.Stopping || cmd == nil {
		t.Fatalf("the run that only holds the list should be stopped first: stopping %v, notice %q", m.life.Stopping, m.notice)
	}
	next, cmd = m.Update(cmd())
	m = next.(Model)
	if !fake.closed || m.running() {
		t.Errorf("the run should be closed: closed %v running %v", fake.closed, m.running())
	}
	if m.session.Path() != path {
		t.Errorf("the list was not opened after the run stopped: %q (%s)", m.session.Path(), m.notice)
	}
	_ = cmd
}

func TestAListOpensWhileTheDownloadsAreStopped(t *testing.T) {
	m, fake := started(t)
	m = download(m)
	m, _ = choose(m, entryPause)
	if !fake.paused {
		t.Fatal("the run did not pause")
	}
	path := otherList(t)

	next, cmd := m.Update(OpenListMsg{Path: path})
	m = next.(Model)
	if !m.life.Stopping || cmd == nil {
		t.Fatalf("stopped downloads should not block opening another list: notice %q", m.notice)
	}
	next, _ = m.Update(cmd())
	m = next.(Model)
	if m.session.Path() != path {
		t.Errorf("the list was not opened: %q (%s)", m.session.Path(), m.notice)
	}
}

func TestAListIsNotOpenedOverSongsThatAreComingIn(t *testing.T) {
	m, fake := started(t)
	m = download(m)

	next, cmd := m.Update(OpenListMsg{Path: otherList(t)})
	m = next.(Model)
	if m.life.Stopping || cmd != nil || fake.closed {
		t.Error("a run that is downloading was stopped without being asked")
	}
	if !strings.Contains(m.notice, "Stop the downloads") {
		t.Errorf("notice %q", m.notice)
	}
}

func TestTheOpenAndNewEntriesStopARunThatIsOnlyHoldingTheList(t *testing.T) {
	for _, entry := range []string{entryOpen, entryNew} {
		m, _ := started(t)
		m, cmd := choose(m, entry)
		if !m.life.Stopping || cmd == nil {
			t.Errorf("%s: the held run should be stopped, notice %q", entry, m.notice)
		}
	}
}
