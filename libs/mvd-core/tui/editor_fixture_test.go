package tui

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	editorpage "youtube-downloader/libs/mvd-core/editor/editor"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

// fakeController stands in for the engine.
type fakeController struct {
	events  chan interface{}
	sources []engine.PlaylistSource
}

func (f *fakeController) Events() <-chan interface{}       { return f.events }
func (f *fakeController) Sources() []engine.PlaylistSource { return f.sources }
func (f *fakeController) LogPath() string                  { return "mvd.log" }
func (f *fakeController) Retry(int) bool                   { return true }
func (f *fakeController) RetryPlaylist(int) int            { return 2 }

// fakePlanRun is a plan-only run that has already decided what to propose.
type fakePlanRun struct {
	fakeController
	plan []engine.PlannedEntry
}

func (r *fakePlanRun) Close()                      {}
func (r *fakePlanRun) Plan() []engine.PlannedEntry { return r.plan }

func proposal(id, title string, index int, target string, kind engine.PlanKind) engine.PlannedEntry {
	return engine.PlannedEntry{
		Playlist: "Mine", PlaylistURL: "https://a",
		Upload:   ytdlp.PlaylistEntry{ID: id, Title: title, Channel: "Band", PlaylistTitle: "Mine", PlaylistIndex: index},
		TargetID: target, Kind: kind, Reason: "because",
	}
}

func samePlan() []engine.PlannedEntry {
	return []engine.PlannedEntry{
		proposal("up00000001", "One (Radio Edit)", 1, "OFFICIAL001", engine.KindOfficial),
		proposal("up00000002", "Two", 2, "BETTER00002", engine.KindBetter),
		proposal("up00000003", "Three", 3, "up00000003", engine.KindOriginal),
	}
}

// editorFixture drives the playlist editor on its own, planning a list with a fake run.
type editorFixture struct {
	t      *testing.T
	editor editorpage.Model
	dir    string
}

func newEditorFixture(t *testing.T, plan []engine.PlannedEntry) *editorFixture {
	t.Helper()
	f := &editorFixture{t: t, dir: t.TempDir()}
	run := &fakePlanRun{fakeController: fakeController{events: make(chan interface{})}, plan: plan}
	host := editorpage.EditorHost{
		Plan: func(config.Config, []string) (Run, error) { return run, nil },
		PlanDownload: func(config.Config, []engine.PlannedEntry) (Run, error) {
			return &fakePlanRun{fakeController: fakeController{events: make(chan interface{})}}, nil
		},
		Open:     func(string) error { return nil },
		Describe: func(id string) (string, error) { return "Title of " + id, nil },
		PickFile: func(string) (string, bool, error) {
			return "", false, errors.New("no file window here")
		},
		SaveFile: func(string, string) (string, bool, error) {
			return "", false, errors.New("no file window here")
		},
		SessionFile: filepath.Join(f.dir, "plan.mvd"),
		SaveDir:     f.dir,
		DecisionLog: filepath.Join(f.dir, "decisions.jsonl"),
	}
	f.editor = editorpage.NewPlanning(host, run, nil, nil)
	f.drive(tea.WindowSizeMsg{Width: 120, Height: 30})
	f.run(f.editor.Init())
	return f
}

// drive feeds msg to the editor and runs the commands it returns that finish at once,
// feeding their results back, the way the Bubble Tea loop would. A command that blocks
// (a wait on the engine's events, a timer) is skipped.
func (f *editorFixture) drive(msg tea.Msg) {
	f.t.Helper()
	next, cmd := f.editor.Update(msg)
	f.editor = next
	f.run(cmd)
}

func (f *editorFixture) run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case tea.BatchMsg:
			for _, c := range msg {
				f.run(c)
			}
		case editorpage.PlannedMsg, editorpage.LeaveMsg, editorpage.DownloadMsg, editorpage.DescribedMsg, editorpage.FilePickedMsg:
			f.drive(msg)
		}
	case <-time.After(30 * time.Millisecond):
	}
}

func (f *editorFixture) press(keys ...string) {
	f.t.Helper()
	for _, k := range keys {
		f.drive(key(k))
	}
}

// skipNote leaves the question about why a decision was changed, which only a build made
// for testing asks.
func (f *editorFixture) skipNote() {
	f.t.Helper()
	if f.editor.Mode() == editorpage.Input {
		f.press("enter")
	}
}

// planDone tells the editor the plan run has dealt with every song.
func (f *editorFixture) planDone() {
	f.t.Helper()
	f.drive(editorpage.EvMsg{Ev: engine.EvIdle{}})
}

func (f *editorFixture) screen() string { return f.editor.Render() }
