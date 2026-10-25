package preferencessession

import (
	"errors"
	"testing"

	"youtube-downloader/libs/mvd-core/settings"
)

func TestTheMoveQuestionsAreAskedAndTheAnswerIsTold(t *testing.T) {
	cases := []struct {
		name    string
		outcome MoveOutcome
		err     error
		want    string
	}{
		{"started", MoveStarted, nil, "MVD was moved to its own folder. This window will close and the new copy will open."},
		{"declined", MoveDeclined, nil, "MVD was not moved."},
		{"already there", MoveAlreadyThere, nil, "MVD already lives in its own folder."},
		{"unavailable", MoveUnavailable, nil, "This machine has no place to move MVD to, or no way to ask. Start MVD with -move from a terminal."},
		{"failed", MoveDeclined, errors.New("disk full"), "MVD could not be moved: disk full"},
	}
	for _, c := range cases {
		s := newSession()
		s.CanMove = true
		if eff := s.AskToMove(); eff.Kind != AskMove || s.Mode != settings.EditConfirming || s.Status != waitingForAnswersNotice {
			t.Fatalf("%s: the questions should be open: %+v mode %v status %q", c.name, eff, s.Mode, s.Status)
		}
		s.MoveAnswered(c.outcome, c.err)
		if s.Mode != settings.EditNone || s.Status != c.want {
			t.Errorf("%s: mode %v, status %q, want %q", c.name, s.Mode, s.Status, c.want)
		}
	}
}

func TestTheUninstallQuestionsAreAskedAndTheAnswerIsTold(t *testing.T) {
	cases := []struct {
		name    string
		outcome RemovalOutcome
		err     error
		want    string
	}{
		{"started", RemovalStarted, nil, "Removing MVD. This window will close."},
		{"declined", RemovalDeclined, nil, "Nothing was removed."},
		{"unavailable", RemovalUnavailable, nil, "This machine cannot ask for confirmation, so nothing was removed. Start MVD with -uninstall from a terminal."},
		{"failed", RemovalDeclined, errors.New("no desktop"), "The removal could not be asked: no desktop"},
	}
	for _, c := range cases {
		s := newSession()
		s.CanUninstall = true
		if eff := s.AskToUninstall(); eff.Kind != AskUninstall || s.Mode != settings.EditConfirming {
			t.Fatalf("%s: the questions should be open: %+v mode %v", c.name, eff, s.Mode)
		}
		s.UninstallAnswered(c.outcome, c.err)
		if s.Mode != settings.EditNone || s.Status != c.want {
			t.Errorf("%s: mode %v, status %q, want %q", c.name, s.Mode, s.Status, c.want)
		}
	}
}

func TestWithoutWhatTheHostCanDoNothingIsAsked(t *testing.T) {
	s := newSession()
	if eff := s.AskToMove(); eff.Kind != NoEffect || s.Mode != settings.EditNone {
		t.Errorf("a host that cannot move the app was asked: %+v mode %v", eff, s.Mode)
	}
	if eff := s.AskToUninstall(); eff.Kind != NoEffect || s.Mode != settings.EditNone {
		t.Errorf("a host that cannot remove the app was asked: %+v mode %v", eff, s.Mode)
	}
}

func TestTheStatusIsForgottenOnRequest(t *testing.T) {
	s := newSession()
	s.Status = "something"
	s.ClearStatus()
	if s.Status != "" {
		t.Errorf("status %q", s.Status)
	}
}
