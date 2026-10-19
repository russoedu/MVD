package workspace

import (
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
)

// fakeRun is a run that does nothing but remember what it was told.
type fakeRun struct {
	events  chan interface{}
	paused  bool
	retried int
	closed  bool
	hold    bool
}

func newFakeRun() *fakeRun                          { return &fakeRun{events: make(chan interface{})} }
func (f *fakeRun) Events() <-chan interface{}       { return f.events }
func (f *fakeRun) Sources() []engine.PlaylistSource { return nil }
func (f *fakeRun) LogPath() string                  { return "" }
func (f *fakeRun) Retry(int) bool                   { return false }
func (f *fakeRun) RetryPlaylist(int) int            { return 0 }
func (f *fakeRun) Close()                           { f.closed = true; close(f.events) }
func (f *fakeRun) Snapshot() []engine.EntrySnapshot { return nil }
func (f *fakeRun) Release(int, string) bool         { return true }
func (f *fakeRun) ReleaseAll() int                  { return 0 }
func (f *fakeRun) Pause()                           { f.paused = true }
func (f *fakeRun) Resume()                          { f.paused = false }
func (f *fakeRun) Paused() bool                     { return f.paused }
func (f *fakeRun) RetryAll() int                    { f.retried++; return 2 }
func (f *fakeRun) Reconfigure(engine.LiveSettings)  {}

func started(t *testing.T, review bool) (Model, *fakeRun) {
	t.Helper()
	fake := newFakeRun()
	host := Host{
		ListDir: t.TempDir(),
		Start: func(_ config.Config, sources []string, _ []engine.PlannedEntry, hold bool) (ListRun, error) {
			fake.hold = hold
			return fake, nil
		},
	}
	list := downloadlist.New([]string{"https://youtu.be/x"})
	if err := list.Save(filepath.Join(t.TempDir(), "a.mvd")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{OutputDir: t.TempDir()}
	m := sized(New(host, cfg, list))
	key := "s"
	if review {
		key = "g"
	}
	m, cmd := press(m, key)
	if cmd == nil || !m.starting {
		t.Fatal("the run did not begin to start")
	}
	next, _ := m.Update(cmd())
	return next.(Model), fake
}

func TestStartingRunsTheListAndPauseAndRetryAllReachTheRun(t *testing.T) {
	m, fake := started(t, false)
	if !m.running() || fake.hold {
		t.Fatalf("running %v hold %v", m.running(), fake.hold)
	}
	m, _ = press(m, "p")
	if !fake.paused {
		t.Error("not paused")
	}
	m, _ = press(m, "p")
	if fake.paused {
		t.Error("not resumed")
	}
	m, _ = press(m, "R")
	if fake.retried != 1 || m.notice != "2 songs queued again" {
		t.Errorf("retried %d, notice %q", fake.retried, m.notice)
	}
}

func TestGettingDataForReviewHoldsTheSongs(t *testing.T) {
	m, fake := started(t, true)
	if !fake.hold || !m.reviewing {
		t.Errorf("hold %v reviewing %v", fake.hold, m.reviewing)
	}
}

func TestStoppingClosesTheRunAndKeepsTheList(t *testing.T) {
	m, fake := started(t, false)
	m, cmd := press(m, "x")
	if !m.stopping || cmd == nil {
		t.Fatal("not stopping")
	}
	msg := cmd()
	next, _ := m.Update(msg)
	m = next.(Model)
	if !fake.closed || m.running() || m.stopping {
		t.Errorf("closed %v running %v stopping %v", fake.closed, m.running(), m.stopping)
	}
	if _, ok := msg.(runClosedMsg); !ok {
		t.Errorf("msg %T", msg)
	}
}

func TestAListDoubleClickedWhileTheAppIsOpenIsOpenedUnlessDownloadsAreGoing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "other.mvd")
	other := downloadlist.New([]string{"https://youtu.be/other"})
	if err := other.Save(path); err != nil {
		t.Fatal(err)
	}

	m := sized(New(Host{}, config.Config{}, nil))
	next, _ := m.Update(OpenListMsg{Path: path})
	m = next.(Model)
	if m.session.Path() != path {
		t.Errorf("the list was not opened: %q (%s)", m.session.Path(), m.notice)
	}

	running, _ := started(t, false)
	next, _ = running.Update(OpenListMsg{Path: path})
	running = next.(Model)
	if running.session.Path() == path {
		t.Error("a list was opened over a run that is going")
	}
	if !strings.Contains(running.notice, "Stop the downloads") {
		t.Errorf("notice %q", running.notice)
	}
}
