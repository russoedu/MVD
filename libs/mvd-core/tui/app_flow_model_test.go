package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
)

type fakeRun struct {
	fakeController
	closed bool
}

func (r *fakeRun) Close() { r.closed = true }

func newFakeRun() *fakeRun {
	return &fakeRun{fakeController: fakeController{
		events:  make(chan interface{}),
		sources: []engine.PlaylistSource{{Index: 0, URL: "https://a"}},
	}}
}

// drive feeds msg to the model and runs the commands it returns that finish
// at once, feeding their results back, the way the Bubble Tea loop would. A
// command that blocks (a wait on the engine's events, a timer) is skipped.
func drive(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	m, cmd := m.Update(msg)
	for _, next := range runNow(cmd) {
		m = drive(t, m, next)
	}
	return m
}

func runNow(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case tea.BatchMsg:
			var out []tea.Msg
			for _, c := range msg {
				out = append(out, runNow(c)...)
			}
			return out
		case setupFinishedMsg, downloadFinishedMsg, runClosedMsg:
			return []tea.Msg{msg}
		}
	case <-time.After(30 * time.Millisecond):
	}
	return nil
}

func appForTest(t *testing.T, urls []string, start RunStarter) (tea.Model, string) {
	t.Helper()
	dir := t.TempDir()
	m := NewAppModel(AppInput{
		Setup: SetupInput{
			Cfg: config.Default(dir, dir), URLs: urls,
			CfgPath: filepath.Join(dir, "config.conf"), ListPath: filepath.Join(dir, "list.txt"),
		},
		Start: start,
	})
	return drive(t, m, tea.WindowSizeMsg{Width: 100, Height: 30}), dir
}

func TestAppModelRunsSetupThenDownloadThenSetup(t *testing.T) {
	run := newFakeRun()
	var started []string
	m, dir := appForTest(t, []string{"https://a"}, func(_ config.Config, urls []string) (Run, error) {
		started = urls
		return run, nil
	})
	if got := m.(appModel).Outline().Title; got != "MVD · Download list (1 item)" {
		t.Fatalf("starts on the list, got %q", got)
	}

	m = drive(t, m, key("ctrl+s"))
	if len(started) != 1 || started[0] != "https://a" {
		t.Fatalf("starter got %v", started)
	}
	if !m.(appModel).downloading {
		t.Fatal("should be on the download screen")
	}

	// Everything finishes, then the user leaves the screen.
	app := m.(appModel)
	app.download.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, Title: "One"}}})
	app.download.state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone})
	app.download.state.Apply(engine.EvIdle{})
	if err := os.WriteFile(filepath.Join(dir, "list.txt"), []byte("https://a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = drive(t, app, key("q"))

	if !run.closed {
		t.Error("the run should be closed")
	}
	back := m.(appModel)
	if back.downloading || back.stopping {
		t.Fatalf("should be back on setup: %+v", back)
	}
	if len(back.urls) != 0 {
		t.Errorf("a finished list should be emptied, got %v", back.urls)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "list.txt")); len(b) != 0 {
		t.Errorf("the saved list should be cleared, got %q", b)
	}
}

func TestAppModelKeepsTheListWhenTheRunIsUnfinished(t *testing.T) {
	run := newFakeRun()
	m, _ := appForTest(t, []string{"https://a"}, func(config.Config, []string) (Run, error) { return run, nil })
	m = drive(t, m, key("ctrl+s"))

	app := m.(appModel)
	app.download.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, Title: "One"}}})
	app.download.state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDownloading})
	m = drive(t, app, tea.KeyMsg{Type: tea.KeyCtrlC})

	back := m.(appModel)
	if back.downloading || !run.closed {
		t.Fatalf("should have stopped the run and left the screen: %+v", back)
	}
	if len(back.urls) != 1 {
		t.Errorf("an unfinished list is kept, got %v", back.urls)
	}
}

func TestAppModelShowsWhyTheRunCouldNotStart(t *testing.T) {
	m, _ := appForTest(t, []string{"https://a"}, func(config.Config, []string) (Run, error) {
		return nil, errors.New("no yt-dlp")
	})
	m = drive(t, m, key("ctrl+s"))

	app := m.(appModel)
	if app.downloading {
		t.Fatal("should stay on setup")
	}
	out := app.Outline()
	if out.Prompt != "Cannot start the download: no yt-dlp" {
		t.Errorf("prompt: %q", out.Prompt)
	}
	m = drive(t, m, key("down"))
	if got := m.(appModel).notice; got != "" {
		t.Errorf("a key press clears the notice, got %q", got)
	}
}

func TestAppModelQuitsFromTheSetupScreens(t *testing.T) {
	m, _ := appForTest(t, nil, func(config.Config, []string) (Run, error) { return newFakeRun(), nil })
	_, cmd := m.Update(key("ctrl+q"))
	if cmd == nil {
		t.Fatal("expected a command")
	}
	next := cmd()
	if _, ok := next.(setupFinishedMsg); !ok {
		t.Fatalf("setup ends with setupFinishedMsg, got %T", next)
	}
	_, cmd = m.Update(next)
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("quitting the setup screens quits the app")
	}
}

func TestDownloadOutlineWindowsALongList(t *testing.T) {
	m := sampleModel(t)
	m.focus = paneEntries
	m.selEntry = 4
	out := m.Outline()
	if len(out.Items) != 6 || !out.Items[4].Selected {
		t.Fatalf("items: %+v", out.Items)
	}
	if out.Items[4].Value != "downloading 34%" || out.Items[2].Value != "failed: Video unavailable" {
		t.Errorf("values: %q, %q", out.Items[4].Value, out.Items[2].Value)
	}
	if len(out.Keys) == 0 {
		t.Error("expected the key bar")
	}

	var entries []engine.EntryInfo
	for i := 0; i < 500; i++ {
		entries = append(entries, engine.EntryInfo{ID: 100 + i, Playlist: 0, Index: i + 1, Title: "t"})
	}
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Big", Entries: entries})
	m.selEntry = 250
	if n := len(m.Outline().Items); n != 2*outlineEntryWindow+1 {
		t.Errorf("want a window of %d, got %d", 2*outlineEntryWindow+1, n)
	}
}
