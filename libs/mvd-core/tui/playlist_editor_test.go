package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

// fakePlanRun is a plan-only run that has already decided what to propose.
type fakePlanRun struct {
	fakeRun
	plan []engine.PlannedEntry
}

func (r *fakePlanRun) Plan() []engine.PlannedEntry { return r.plan }

func proposal(id, title string, index int, target string, kind engine.PlanKind) engine.PlannedEntry {
	return engine.PlannedEntry{
		Playlist: "Mine", PlaylistURL: "https://a",
		Upload:   ytdlp.PlaylistEntry{ID: id, Title: title, Channel: "Band", PlaylistTitle: "Mine", PlaylistIndex: index},
		TargetID: target, Kind: kind, Reason: "because",
	}
}

type editorFixture struct {
	t         *testing.T
	m         tea.Model
	dir       string
	planRun   *fakePlanRun
	planned   []string
	download  *fakeRun
	handed    []engine.PlannedEntry
	described []string
	opened    []string
}

func newEditorFixture(t *testing.T, urls []string, plan []engine.PlannedEntry) *editorFixture {
	t.Helper()
	f := &editorFixture{t: t, dir: t.TempDir(), download: newFakeRun()}
	f.planRun = &fakePlanRun{fakeRun: *newFakeRun(), plan: plan}
	host := EditorHost{
		Plan: func(_ config.Config, urls []string) (Run, error) {
			f.planned = urls
			return f.planRun, nil
		},
		PlanDownload: func(_ config.Config, plan []engine.PlannedEntry) (Run, error) {
			f.handed = plan
			return f.download, nil
		},
		Open:        func(url string) error { f.opened = append(f.opened, url); return nil },
		Describe:    func(id string) (string, error) { f.described = append(f.described, id); return "Title of " + id, nil },
		SessionFile: filepath.Join(f.dir, "plan.mvd"),
		SaveDir:     f.dir,
		DecisionLog: filepath.Join(f.dir, "decisions.jsonl"),
	}
	m := NewAppModel(AppInput{
		Setup: SetupInput{
			Cfg: config.Default(f.dir, f.dir), URLs: urls,
			CfgPath: filepath.Join(f.dir, "config.conf"), ListPath: filepath.Join(f.dir, "list.txt"),
		},
		Start:  func(config.Config, []string) (Run, error) { return newFakeRun(), nil },
		Editor: host,
	})
	f.m = drive(t, m, tea.WindowSizeMsg{Width: 120, Height: 30})
	return f
}

func (f *editorFixture) press(keys ...string) {
	f.t.Helper()
	for _, k := range keys {
		if k == "ctrl+e" {
			f.m = drive(f.t, f.m, tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
			continue
		}
		f.m = drive(f.t, f.m, key(k))
	}
}

func (f *editorFixture) app() appModel { return f.m.(appModel) }

// skipNote leaves the question about why a decision was changed, which only a build made
// for testing asks.
func (f *editorFixture) skipNote() {
	f.t.Helper()
	if f.app().editor.mode == editorInput {
		f.m = drive(f.t, f.m, key("enter"))
	}
}

// planDone tells the editor the plan run has dealt with every song.
func (f *editorFixture) planDone() {
	f.t.Helper()
	f.m = drive(f.t, f.m, editorEvMsg{engine.EvIdle{}})
}

func (f *editorFixture) screen() string { return f.app().render() }

func samePlan() []engine.PlannedEntry {
	return []engine.PlannedEntry{
		proposal("up00000001", "One (Radio Edit)", 1, "OFFICIAL001", engine.KindOfficial),
		proposal("up00000002", "Two", 2, "BETTER00002", engine.KindBetter),
		proposal("up00000003", "Three", 3, "up00000003", engine.KindOriginal),
	}
}

func TestReviewPlansTheListThenDownloadsWhatWasConfirmedOrReplaced(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())

	f.press("ctrl+e")
	app := f.app()
	if !app.editing || !app.editor.planning || len(f.planned) != 1 || f.planned[0] != "https://a" {
		t.Fatalf("a review should plan the list, got editing=%v planning=%v planned=%v", app.editing, app.editor.planning, f.planned)
	}
	if !strings.Contains(f.screen(), "Looking up the best version") {
		t.Errorf("the planning screen:\n%s", f.screen())
	}

	f.planDone()
	if !f.planRun.closed {
		t.Error("the plan run should be closed once the plan is read")
	}
	app = f.app()
	if app.editor.planning || len(app.editor.rows) != 3 {
		t.Fatalf("rows %+v", app.editor.rows)
	}
	if screen := f.screen(); !strings.Contains(screen, "One (Radio Edit)") || !strings.Contains(screen, "OFFICIAL001") || !strings.Contains(screen, "3 songs") {
		t.Errorf("the songs and their proposals should be listed:\n%s", screen)
	}

	// Confirm the first (moves on to the second), replace the second by a video of their own.
	f.press("c", "r")
	f.m = drive(t, f.m, tea.PasteMsg{Content: "https://youtu.be/MINE0000001"})
	f.m = drive(t, f.m, key("enter"))
	f.skipNote()
	rows := f.app().editor.rows
	if rows[0].Decision != playlistfile.DecisionConfirmed || rows[1].Decision != playlistfile.DecisionReplaced || rows[1].ChosenID != "MINE0000001" {
		t.Fatalf("decisions %+v", rows)
	}
	if len(f.described) != 1 || f.described[0] != "MINE0000001" || rows[1].ChosenTitle != "Title of MINE0000001" {
		t.Errorf("the video they picked should be named: %v, %q", f.described, rows[1].ChosenTitle)
	}

	// The third was not looked at: the person is asked what to do with it.
	f.press("d")
	if f.app().editor.mode != editorAskDownload || !strings.Contains(f.screen(), "1 songs were not reviewed") {
		t.Fatalf("expected the question about unreviewed songs:\n%s", f.screen())
	}
	f.press("y")
	if len(f.handed) != 2 || !f.app().downloading {
		t.Fatalf("the download should start with 2 songs, handed %+v", f.handed)
	}
	if f.handed[0].TargetID != "OFFICIAL001" || f.handed[1].TargetID != "MINE0000001" || f.handed[1].Kind != engine.KindChosen || f.handed[1].Upload.ID != "up00000002" {
		t.Errorf("handed %+v", f.handed)
	}

	// Both finish; the person leaves the download screen and is back on the editor.
	app = f.app()
	app.download.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Mine", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, Title: "One"}, {ID: 1, Playlist: 0, Index: 2, Title: "Two"}}})
	app.download.state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone})
	app.download.state.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDone})
	app.download.state.Apply(engine.EvIdle{})
	f.m = drive(t, app, key("q"))

	back := f.app()
	if back.downloading || !back.editing || !f.download.closed {
		t.Fatalf("should be back on the editor: downloading=%v editing=%v closed=%v", back.downloading, back.editing, f.download.closed)
	}
	if !back.editor.rows[0].Downloaded || !back.editor.rows[1].Downloaded || back.editor.rows[2].Downloaded {
		t.Errorf("downloaded flags %+v", back.editor.rows)
	}
	saved, err := playlistfile.Load(filepath.Join(f.dir, "plan.mvd"))
	if err != nil || len(saved.Entries) != 3 || !saved.Entries[0].Downloaded || saved.Entries[1].ChosenID != "MINE0000001" {
		t.Fatalf("the session file: %+v, %v", saved, err)
	}

	// What was downloaded is skipped from now on: only the third is left, and as it was
	// not looked at, taking it as proposed is a choice.
	f.press("d", "a")
	if len(f.handed) != 1 || f.handed[0].Upload.ID != "up00000003" {
		t.Errorf("only the song left should be handed, got %+v", f.handed)
	}
}

func TestAVideoIsOpenedInTheBrowserFromTheEditor(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	f.press("ctrl+e")
	f.planDone()

	// "o" opens what would be downloaded, "v" the upload that is in the playlist.
	f.press("o", "down", "v")
	want := []string{"https://www.youtube.com/watch?v=OFFICIAL001", "https://www.youtube.com/watch?v=up00000002"}
	if strings.Join(f.opened, " ") != strings.Join(want, " ") {
		t.Errorf("opened %v, want %v", f.opened, want)
	}
}

func TestAReviewFromLastTimeIsOfferedAndBroughtAlong(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	last := []playlistfile.Entry{
		{Playlist: "Mine", PlaylistURL: "https://a", Upload: ytdlp.PlaylistEntry{ID: "up00000001", Title: "One (Radio Edit)", PlaylistIndex: 1}, TargetID: "OFFICIAL001", Kind: "official", Decision: playlistfile.DecisionConfirmed, Downloaded: true},
		{Playlist: "Mine", PlaylistURL: "https://a", Upload: ytdlp.PlaylistEntry{ID: "up00000002", Title: "Two", PlaylistIndex: 2}, TargetID: "BETTER00002", Kind: "better", Decision: playlistfile.DecisionSkipped},
	}
	if err := playlistfile.Save(filepath.Join(f.dir, "plan.mvd"), playlistfile.File{Entries: last}); err != nil {
		t.Fatal(err)
	}

	f.press("ctrl+e")
	if f.app().editor.mode != editorAskPrevious || !strings.Contains(f.screen(), "from last time") {
		t.Fatalf("the question should be on the screen:\n%s", f.screen())
	}
	// The plan can finish while the question waits.
	f.planDone()
	if !f.app().editor.planning {
		t.Fatal("the plan should wait for the answer")
	}
	f.press("y")

	rows := f.app().editor.rows
	if f.app().editor.planning || len(rows) != 3 {
		t.Fatalf("rows %+v", rows)
	}
	if !rows[0].Downloaded || rows[1].Decision != playlistfile.DecisionSkipped || rows[2].Decision != "" {
		t.Errorf("the decisions of last time should be kept: %+v", rows)
	}
}

func TestAReviewFromLastTimeCanBeLeftBehind(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	last := []playlistfile.Entry{{Playlist: "Mine", PlaylistURL: "https://a", Upload: ytdlp.PlaylistEntry{ID: "up00000001"}, TargetID: "X", Kind: "official", Decision: playlistfile.DecisionSkipped}}
	if err := playlistfile.Save(filepath.Join(f.dir, "plan.mvd"), playlistfile.File{Entries: last}); err != nil {
		t.Fatal(err)
	}
	f.press("ctrl+e", "n")
	f.planDone()
	for _, r := range f.app().editor.rows {
		if r.Decision != "" {
			t.Errorf("nothing should be carried over: %+v", r)
		}
	}
}

func TestASavedPlanInTheListOpensTheEditorWithoutPlanning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mine.mvd")
	rows := []playlistfile.Entry{
		{Playlist: "Mine", PlaylistURL: "https://a", Upload: ytdlp.PlaylistEntry{ID: "up00000001", Title: "One"}, TargetID: "OFFICIAL001", Kind: "official", Decision: playlistfile.DecisionConfirmed},
	}
	if err := playlistfile.Save(path, playlistfile.File{Entries: rows}); err != nil {
		t.Fatal(err)
	}

	f := newEditorFixture(t, []string{path}, nil)
	f.press("ctrl+s")
	app := f.app()
	if !app.editing || app.editor.planning || len(app.editor.rows) != 1 || len(f.planned) != 0 {
		t.Fatalf("editing=%v planning=%v rows=%d planned=%v", app.editing, app.editor.planning, len(app.editor.rows), f.planned)
	}
	f.press("d")
	if len(f.handed) != 1 || f.handed[0].TargetID != "OFFICIAL001" {
		t.Errorf("handed %+v", f.handed)
	}
}

func TestAFileThatIsNotAPlanIsRefusedOnTheSetupScreen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.mvd")
	if err := os.WriteFile(path, []byte("not a plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := newEditorFixture(t, []string{path}, nil)
	f.press("ctrl+s")
	app := f.app()
	if app.editing || !strings.Contains(app.notice, "Cannot open") {
		t.Errorf("editing=%v notice=%q", app.editing, app.notice)
	}
}

func TestLeavingWhilePlanningStopsTheRunAndGoesBackToTheList(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	f.press("ctrl+e", "esc")
	app := f.app()
	if !f.planRun.closed || app.editing || app.stopping {
		t.Fatalf("closed=%v editing=%v stopping=%v", f.planRun.closed, app.editing, app.stopping)
	}
	if len(app.urls) != 1 {
		t.Errorf("the list should be kept, got %v", app.urls)
	}
}

func TestSaveAndOpenAPlanFile(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	f.press("ctrl+e")
	f.planDone()
	f.press("c", "c")

	target := filepath.Join(f.dir, "chunk")
	f.press("s")
	// The prompt starts with the playlist's name in the saved plans folder: replace it.
	app := f.app()
	app.editor.input.SetValue(target)
	f.m = app
	f.m = drive(t, f.m, key("enter"))

	saved, err := playlistfile.Load(target + ".mvd")
	if err != nil || len(saved.Entries) != 3 || saved.Entries[0].Decision != playlistfile.DecisionConfirmed {
		t.Fatalf("saved %+v, %v", saved, err)
	}

	// Open it again after undoing a decision.
	f.press("z")
	f.press("l")
	app = f.app()
	app.editor.input.SetValue(target + ".mvd")
	f.m = app
	f.m = drive(t, f.m, key("enter"))
	if got := f.app().editor.rows[1].Decision; got != playlistfile.DecisionConfirmed {
		t.Errorf("the saved decisions should be back, got %q", got)
	}
}

func TestOnlyAVideoAddressIsTakenAsAReplacement(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	f.press("ctrl+e")
	f.planDone()
	f.press("r")
	f.m = drive(t, f.m, tea.PasteMsg{Content: "https://example.com/not-a-video"})
	f.m = drive(t, f.m, key("enter"))
	if f.app().editor.rows[0].Decision != "" || !strings.Contains(f.screen(), "not the address of a YouTube video") {
		t.Errorf("the decision should be untouched and the reason shown:\n%s", f.screen())
	}
}

func TestASongNothingWasFoundForCannotBeConfirmed(t *testing.T) {
	plan := []engine.PlannedEntry{proposal("up00000001", "Lost", 1, "up00000001", engine.KindNone)}
	f := newEditorFixture(t, []string{"https://a"}, plan)
	f.press("ctrl+e")
	f.planDone()
	f.press("c")
	if f.app().editor.rows[0].Decision != "" {
		t.Error("confirming a song with no proposal should do nothing")
	}
	f.press("u")
	f.skipNote()
	if f.app().editor.rows[0].Decision != playlistfile.DecisionOriginal {
		t.Error("the original of an unfound song can be taken")
	}
}

func TestTheEditorHasAnOutlineForScreenReaders(t *testing.T) {
	f := newEditorFixture(t, []string{"https://a"}, samePlan())
	f.press("ctrl+e")
	if out := f.app().Outline(); !strings.Contains(out.Prompt, "Looking up") {
		t.Errorf("planning outline %+v", out)
	}
	f.planDone()
	out := f.app().Outline()
	if len(out.Items) != 3 || !out.Items[0].Selected || out.ItemsLabel != "Songs" || len(out.Keys) == 0 {
		t.Errorf("outline %+v", out)
	}
}
