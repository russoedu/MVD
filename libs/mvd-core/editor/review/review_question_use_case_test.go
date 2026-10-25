package review

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

func TestDownloadingWhatWasReviewedStartsAtOnce(t *testing.T) {
	rows := []playlistfile.Entry{song("u1", "o1", engine.KindOfficial)}
	rows[0].Decision = playlistfile.DecisionConfirmed
	s := NewFromRows(Options{CanDownload: true}, rows)
	effects := s.Download()
	if len(effects) != 1 {
		t.Fatalf("effects %v", effects)
	}
	d, ok := effects[0].(Download)
	if !ok || len(d.Plan) != 1 || d.Plan[0].TargetID != "o1" || len(d.Keys) != 1 {
		t.Fatalf("effect %#v", effects[0])
	}
}

func TestSongsNobodyLookedAtAreAskedAboutBeforeDownloading(t *testing.T) {
	s := NewFromRows(Options{CanDownload: true}, threeSongs())
	s.Confirm()
	if effects := s.Download(); len(effects) != 0 || s.Mode != AskDownload {
		t.Fatalf("effects %v mode %v", effects, s.Mode)
	}
	if got := s.DownloadQuestion(); got != "1 songs were not reviewed. " {
		t.Errorf("question %q", got)
	}

	if effects := s.AnswerDownload(DownloadBack); len(effects) != 0 || s.Mode != Browsing {
		t.Fatalf("back: effects %v mode %v", effects, s.Mode)
	}

	s.Download()
	effects := s.AnswerDownload(DownloadAsProposed)
	if len(effects) != 1 || len(effects[0].(Download).Plan) != 2 {
		t.Fatalf("as proposed: %#v", effects)
	}
	s.Download()
	effects = s.AnswerDownload(DownloadReviewed)
	if len(effects) != 1 || len(effects[0].(Download).Plan) != 1 {
		t.Fatalf("reviewed only: %#v", effects)
	}
}

func TestNothingToDownloadSaysSo(t *testing.T) {
	s := NewFromRows(Options{CanDownload: true}, []playlistfile.Entry{song("u1", "", engine.KindNone)})
	s.Mode = AskDownload
	s.AnswerDownload(DownloadReviewed)
	if text, _ := s.NoticeShown(time.Now()); text == "" {
		t.Error("no notice")
	}
}

func TestADownloadThePlaceCannotDoIsRefused(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	if effects := s.Download(); len(effects) != 0 {
		t.Fatalf("effects %v", effects)
	}
	if text, _ := s.NoticeShown(time.Now()); text != "downloading is not available here" {
		t.Errorf("notice %q", text)
	}
}

func TestMarkingSongsDownloadedSkipsThemNextTime(t *testing.T) {
	s := NewFromRows(Options{}, threeSongs())
	s.MarkDownloaded([]string{s.Rows[0].Key()})
	if !s.Rows[0].Downloaded {
		t.Fatal("not marked")
	}
	if text, _ := s.NoticeShown(time.Now()); text != "1 song downloaded; it will be skipped next time" {
		t.Errorf("notice %q", text)
	}
	s.MarkDownloaded(nil)
	if text, _ := s.NoticeShown(time.Now()); text != "nothing was downloaded" {
		t.Errorf("notice %q", text)
	}
}

func TestSavingAsksForANameOnTheScreenWhenTheHostHasNoWindow(t *testing.T) {
	dir := t.TempDir()
	rows := threeSongs()
	rows[0].Playlist = `my: list`
	s := NewFromRows(Options{SaveDir: dir}, rows)
	if effects := s.Save(); len(effects) != 0 || s.Mode != Input {
		t.Fatalf("effects %v mode %v", effects, s.Mode)
	}
	if want := filepath.Join(dir, "my_ list"+playlistfile.Extension); s.InputText != want {
		t.Fatalf("offered %q, want %q", s.InputText, want)
	}
	s.SetInput(filepath.Join(dir, "review"))
	s.SubmitInput()
	loaded, err := playlistfile.Load(filepath.Join(dir, "review"+playlistfile.Extension))
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Songs()) != 3 {
		t.Errorf("saved %d songs", len(loaded.Songs()))
	}

	other := NewFromRows(Options{SaveDir: dir}, nil)
	other.Load()
	other.SetInput(filepath.Join(dir, "review"+playlistfile.Extension))
	other.SubmitInput()
	if len(other.Rows) != 3 || other.Selected != 0 {
		t.Fatalf("opened %d songs", len(other.Rows))
	}
}

func TestSavingAndOpeningUseTheWindowOfTheHostWhenItHasOne(t *testing.T) {
	dir := t.TempDir()
	s := NewFromRows(Options{SaveDir: dir, CanSaveFile: true, CanPickFile: true}, threeSongs())

	effects := s.Save()
	save, ok := effects[0].(PickSaveFile)
	if !ok || save.Dir != filepath.Clean(dir) || !s.Picking {
		t.Fatalf("effect %#v picking %v", effects[0], s.Picking)
	}
	s.FilePicked("", false, nil, true)
	if s.Picking || s.Mode != Browsing {
		t.Fatalf("a cancelled window left picking %v mode %v", s.Picking, s.Mode)
	}

	effects = s.Load()
	if open, ok := effects[0].(PickOpenFile); !ok || open.Start != dir {
		t.Fatalf("effect %#v", effects[0])
	}
	s.FilePicked("", false, errors.New("no window"), false)
	if s.Mode != Input || s.Prompt != "Open the file: " {
		t.Errorf("a window that could not be shown should ask on the screen: mode %v prompt %q", s.Mode, s.Prompt)
	}
}

func TestAReviewThatGoesOnIsSavedWithTheList(t *testing.T) {
	s := NewFromRows(Options{CanDownload: true, CanSaveFile: true}, threeSongs())
	s.live = fakeLive{}
	if s.Save() != nil || s.Load() != nil || s.Download() != nil || s.Mode != Browsing || s.Picking {
		t.Fatal("a live review saved, opened or downloaded by itself")
	}
}

func TestThePreviousReviewIsBroughtAlongOrLeft(t *testing.T) {
	saved := []playlistfile.Entry{song("u1", "o1", engine.KindOfficial), song("u9", "o9", engine.KindOfficial)}
	saved[0].Decision = playlistfile.DecisionConfirmed
	saved[1].Downloaded = true

	run := &fakePlanRun{plan: []engine.PlannedEntry{{PlaylistURL: "https://list", Upload: saved[0].Upload, TargetID: "o1", Kind: engine.KindOfficial}}}
	s := NewPlanning(Options{}, run, nil, saved)
	if s.Mode != AskPrevious || s.PendingCount() != 1 {
		t.Fatalf("mode %v pending %d", s.Mode, s.PendingCount())
	}

	if _, over := s.ApplyEvent(engine.EvIdle{}); !over || !s.Planning {
		t.Fatalf("the plan was taken while the question waited: over %v planning %v", over, s.Planning)
	}
	effects := s.AnswerPrevious(true)
	if s.Planning || len(effects) != 1 || effects[0] != (PlanFinished{}) {
		t.Fatalf("planning %v effects %v", s.Planning, effects)
	}
	if len(s.Rows) != 2 || s.Rows[0].Decision != playlistfile.DecisionConfirmed {
		t.Fatalf("the saved review was not brought along: %+v", s.Rows)
	}

	left := NewPlanning(Options{}, run, nil, saved)
	left.AnswerPrevious(false)
	left.ApplyEvent(engine.EvIdle{})
	if len(left.Rows) != 1 || left.Rows[0].Decision != playlistfile.DecisionNone {
		t.Fatalf("the saved review was brought along: %+v", left.Rows)
	}
}

func TestAPlanThatFoundNothingSaysSo(t *testing.T) {
	s := NewPlanning(Options{}, &fakePlanRun{}, nil, nil)
	if s.Mode != Browsing {
		t.Fatalf("mode %v", s.Mode)
	}
	s.ApplyEvent(engine.EvIdle{})
	if text, _ := s.NoticeShown(time.Now()); text != "no songs were found in the list" {
		t.Errorf("notice %q", text)
	}
}
