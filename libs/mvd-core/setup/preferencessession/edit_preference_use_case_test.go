package preferencessession

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
)

func newSession() Session { return New(config.Default("/app", "/dl")) }

func TestAChoiceIsMadeAmongTheOptionsAndCommitted(t *testing.T) {
	s := newSession()
	s.Cursor = itemVideoQuality
	start := indexOf(config.VideoPresets, s.Cfg.VideoQuality)
	if eff := s.Activate(); eff.Kind != NoEffect || s.Mode != settings.EditRadio || s.RadioIdx != start {
		t.Fatalf("enter should open the choices at the current one: mode %v idx %d", s.Mode, s.RadioIdx)
	}
	s.MoveChoice(1)
	s.CommitEdit()
	if s.Mode != settings.EditNone || s.Cfg.VideoQuality != config.VideoPresets[start+1] {
		t.Errorf("video quality should be %q, got %q", config.VideoPresets[start+1], s.Cfg.VideoQuality)
	}
}

func TestTheChoiceStaysAmongTheOptions(t *testing.T) {
	s := newSession()
	s.Cursor = itemMergeFormat
	s.Activate()
	s.MoveChoice(-100)
	if s.RadioIdx != 0 {
		t.Errorf("the choice left the top: %d", s.RadioIdx)
	}
	s.MoveChoice(100)
	if s.RadioIdx != len(mergeOptions)-1 {
		t.Errorf("the choice left the bottom: %d", s.RadioIdx)
	}
}

func TestOfficialVideoIsAChoiceAndSaveNotFoundIsASwitch(t *testing.T) {
	s := newSession()
	s.Cursor = itemOfficial
	s.Activate()
	for s.RadioOpts[s.RadioIdx] != string(config.OfficialOnly) {
		s.MoveChoice(1)
	}
	s.CommitEdit()
	if s.Cfg.OfficialVideo != config.OfficialOnly {
		t.Fatalf("the choice should be only, it is %q", s.Cfg.OfficialVideo)
	}
	s.Cursor = itemSaveNotFound
	before := s.Cfg.SaveNotFound
	s.Activate()
	if s.Cfg.SaveNotFound == before || s.Editing() {
		t.Errorf("save-not-found should turn over without editing")
	}
}

func TestANumberStepsUpAndDownButNotBelowOne(t *testing.T) {
	s := newSession()
	s.Cursor = itemConcurrent
	before := s.Cfg.MaxConcurrentDownloads
	s.Activate()
	if s.Mode != settings.EditNumber {
		t.Fatalf("enter should edit the number, mode %v", s.Mode)
	}
	s.StepNumber(1)
	s.CommitEdit()
	if s.Cfg.MaxConcurrentDownloads != before+1 {
		t.Errorf("up then commit should give %d, got %d", before+1, s.Cfg.MaxConcurrentDownloads)
	}
	s.Activate()
	s.SetText("1")
	s.StepNumber(-1)
	if s.Text != "1" {
		t.Errorf("a number stepped below one: %q", s.Text)
	}
	s.SetText("nonsense")
	s.CommitEdit()
	if s.Cfg.MaxConcurrentDownloads != before+1 {
		t.Errorf("a value that is not a number changed the setting to %d", s.Cfg.MaxConcurrentDownloads)
	}
}

func TestATextIsTrimmedCommittedAndNeverEmpty(t *testing.T) {
	s := newSession()
	s.Cursor = itemTemplate
	was := s.Cfg.OutputTemplate
	s.Activate()
	if s.Mode != settings.EditText || s.Text != was {
		t.Fatalf("enter should edit the template from its value: mode %v text %q", s.Mode, s.Text)
	}
	s.SetText("  %(title)s.%(ext)s ")
	s.CommitEdit()
	if s.Cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("text edit should commit trimmed, got %q", s.Cfg.OutputTemplate)
	}
	s.Activate()
	s.SetText("   ")
	s.CommitEdit()
	if s.Cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("an empty template replaced the setting: %q", s.Cfg.OutputTemplate)
	}
}

func TestCancellingAnEditChangesNothing(t *testing.T) {
	s := newSession()
	s.Cursor = itemTemplate
	was := s.Cfg.OutputTemplate
	s.Activate()
	s.SetText("other")
	s.CancelEdit()
	if s.Editing() || s.Cfg.OutputTemplate != was {
		t.Errorf("editing %v template %q", s.Editing(), s.Cfg.OutputTemplate)
	}
}

func TestTheCursorStaysInTheListAndSelectIsRefusedWhileEditing(t *testing.T) {
	s := newSession()
	s.Move(-1)
	s.Move(100)
	if s.Cursor != ItemCount-1 {
		t.Errorf("the cursor left the list: %d", s.Cursor)
	}
	s.Select(2)
	s.Activate()
	s.Select(0)
	if s.Cursor != 2 {
		t.Errorf("the cursor moved by a click while a choice was open: %d", s.Cursor)
	}
}

func TestEachExitAsksTheHostForIt(t *testing.T) {
	s := newSession()
	if s.Keep().Kind != Save || s.Leave().Kind != Cancel || s.OpenAdvanced().Kind != OpenAdvanced || s.OpenColours().Kind != OpenColours {
		t.Error("an exit asked the host for the wrong thing")
	}
}

func TestOnlyTheTemplateHasAHint(t *testing.T) {
	s := newSession()
	if s.OnTemplate() {
		t.Error("the first setting is not the template")
	}
	s.Cursor = itemTemplate
	if !s.OnTemplate() {
		t.Error("the template setting should have its hint")
	}
}

func TestACookieChoiceIsAppliedAndReadBack(t *testing.T) {
	cfg := config.Default("/app", "/dl")
	ApplyCookieChoice(&cfg, "firefox")
	if cfg.CookiesFromBrowser != "firefox" || cfg.AutoCookies || CurrentCookieChoice(cfg) != "firefox" {
		t.Errorf("pin wrong: %+v", cfg)
	}
	ApplyCookieChoice(&cfg, "off")
	if cfg.AutoCookies || cfg.CookiesFromBrowser != "" || CurrentCookieChoice(cfg) != "off" {
		t.Errorf("off wrong: %+v", cfg)
	}
	ApplyCookieChoice(&cfg, "all")
	if !cfg.AutoCookies || CurrentCookieChoice(cfg) != "all" {
		t.Errorf("all wrong: %+v", cfg)
	}
}

func TestSaveNotFoundSaysWhenItIsNotInUse(t *testing.T) {
	s := newSession()
	s.Cfg.OfficialVideo = config.OfficialOnly
	s.Cfg.SaveNotFound = true
	if got := s.Display(itemSaveNotFound); got != "yes" {
		t.Errorf("with official only the switch reads plainly, got %q", got)
	}
	s.Cfg.OfficialVideo = config.OfficialYes
	if got := s.Display(itemSaveNotFound); got != "yes (used when Download Official is only)" {
		t.Errorf("otherwise it says it is not used, got %q", got)
	}
}
