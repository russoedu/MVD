package coloursession

import (
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

func TestAnInvalidColourIsRefusedAndKeepsTheEditorOpen(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	s.BeginEdit()
	s.SetText("purple")
	if eff := s.AcceptEdit(); eff != NoEffect {
		t.Errorf("an invalid colour asked the host for %v", eff)
	}
	if !s.Editing || s.Status != InvalidColourNotice || s.Cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Fatalf("an invalid colour must be refused: editing=%v status=%q accent=%q", s.Editing, s.Status, s.Cfg.Colors.Accent)
	}
}

func TestAValidColourIsAcceptedAndTheHostAppliesIt(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	s.BeginEdit()
	s.SetText(" #123456 ")
	if eff := s.AcceptEdit(); eff != ApplyColours {
		t.Errorf("an accepted colour asked the host for %v", eff)
	}
	if s.Editing || s.Cfg.Colors.Accent != "#123456" || s.Status != "" {
		t.Errorf("a valid colour should be accepted: %+v editing=%v", s.Cfg.Colors, s.Editing)
	}
}

func TestDefaultPutsTheColourOfTheCursorBack(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	s.BeginEdit()
	s.SetText("#123456")
	s.AcceptEdit()
	if eff := s.RestoreDefault(); eff != ApplyColours || s.Cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Errorf("default gave %v and %q", eff, s.Cfg.Colors.Accent)
	}
}

func TestTheCursorStaysInTheListAndStaysPutWhileTyping(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	s.Move(-1)
	if s.Cursor != 0 {
		t.Errorf("the cursor left the top: %d", s.Cursor)
	}
	s.Move(100)
	if s.Cursor != len(Labels)-1 {
		t.Errorf("the cursor left the bottom: %d", s.Cursor)
	}
	s.Select(2)
	s.BeginEdit()
	s.Move(1)
	s.Select(0)
	if s.Cursor != 2 {
		t.Errorf("the cursor moved while a colour was typed: %d", s.Cursor)
	}
}

func TestBeginningToEditStartsFromTheColourItHas(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	s.Status = "old"
	s.BeginEdit()
	if !s.Editing || s.Text != s.Value(0) || s.Status != "" {
		t.Errorf("editing=%v text=%q status=%q", s.Editing, s.Text, s.Status)
	}
	s.CancelEdit()
	if s.Editing {
		t.Error("cancel left the editor open")
	}
}

func TestSavingAndLeavingAskTheHost(t *testing.T) {
	s := New(config.Default("/app", "/dl"))
	if s.Keep() != Save || s.Leave() != Cancel {
		t.Error("keep should save and leave should cancel")
	}
}
