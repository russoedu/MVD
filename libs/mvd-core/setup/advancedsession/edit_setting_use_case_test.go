package advancedsession

import (
	"strconv"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
)

func newSession() Session { return New(config.Default("/app", "/dl")) }

func TestActivatingASwitchTurnsItOver(t *testing.T) {
	s := newSession()
	s.Cursor = itemAutoRetry
	before := s.Cfg.AutoRetry
	s.Activate()
	if s.Cfg.AutoRetry == before || s.Editing() {
		t.Errorf("auto retry should turn over without typing: now %v editing %v", s.Cfg.AutoRetry, s.Editing())
	}
}

func TestANumberStepsDownAndIsCommitted(t *testing.T) {
	s := newSession()
	s.Cursor = itemFragments
	start := s.Cfg.ConcurrentFragments
	s.Activate()
	if s.Mode != settings.EditNumber || s.Text == "" {
		t.Fatalf("activating the fragment count should edit it as a number: mode %v text %q", s.Mode, s.Text)
	}
	s.StepNumber(-1)
	s.CommitEdit()
	if s.Cfg.ConcurrentFragments != start-1 || s.Editing() {
		t.Errorf("fragments should step down to %d, got %d", start-1, s.Cfg.ConcurrentFragments)
	}
}

func TestANumberDoesNotStepBelowZero(t *testing.T) {
	s := newSession()
	s.Cursor = itemFragments
	s.Activate()
	s.SetText("0")
	s.StepNumber(-1)
	if s.Text != "0" {
		t.Errorf("a number stepped below zero: %q", s.Text)
	}
	s.StepNumber(1)
	if s.Text != "1" {
		t.Errorf("a step up gave %q", s.Text)
	}
}

func TestTheRetrySettingsTakeOnlyNumbersFromOne(t *testing.T) {
	s := newSession()
	for cursor, want := range map[int]int{itemRetryCycles: 12, itemRetryWait: 7, itemStartTimeout: 45} {
		s.Cursor = cursor
		s.Activate()
		s.SetText(strconv.Itoa(want))
		s.CommitEdit()
	}
	if s.Cfg.RetryCycles != 12 || s.Cfg.RetryWaitSeconds != 7 || s.Cfg.StartTimeoutSeconds != 45 {
		t.Errorf("retry settings not committed: %d cycles, %d s, %d s", s.Cfg.RetryCycles, s.Cfg.RetryWaitSeconds, s.Cfg.StartTimeoutSeconds)
	}
	s.Cursor = itemRetryWait
	s.Activate()
	s.SetText("0")
	s.CommitEdit()
	if s.Cfg.RetryWaitSeconds != 7 {
		t.Errorf("a wait of 0 should be refused, got %d", s.Cfg.RetryWaitSeconds)
	}
	if got := s.Display(itemRetryWait); !strings.Contains(got, "7s, 14s, 21s") {
		t.Errorf("the wait should show how it grows, got %q", got)
	}
}

func TestExtraArgsAreSplitOnSpaces(t *testing.T) {
	s := newSession()
	s.Activate()
	s.SetText("  --a   --b x ")
	s.CommitEdit()
	if got := s.Display(itemExtraArgs); got != "--a --b x" {
		t.Errorf("extra args read %q", got)
	}
	s.Cursor = itemExtraArgs
	s.Activate()
	s.SetText("")
	s.CommitEdit()
	if got := s.Display(itemExtraArgs); got != "(none)" {
		t.Errorf("no extra args read %q", got)
	}
}

func TestTheCursorStaysInTheList(t *testing.T) {
	s := newSession()
	s.Move(-1)
	if s.Cursor != 0 {
		t.Errorf("the cursor left the top: %d", s.Cursor)
	}
	s.Move(100)
	if s.Cursor != len(Labels)-1 {
		t.Errorf("the cursor left the bottom: %d", s.Cursor)
	}
	s.Cursor = 1
	s.Activate()
	s.Select(0)
	if s.Cursor != 1 {
		t.Error("the cursor moved by a click while a number was typed")
	}
}

func TestSavingAndLeavingAskTheHost(t *testing.T) {
	s := newSession()
	if s.Keep() != Save || s.Leave() != Cancel {
		t.Error("keep should save and leave should cancel")
	}
}
