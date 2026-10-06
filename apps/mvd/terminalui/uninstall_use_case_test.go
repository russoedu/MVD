package terminalui

import (
	"errors"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/uninstall"
	"youtube-downloader/libs/mvd-core/tui"
)

type fakeRemover struct {
	err     error
	removed chan struct{}
}

func (f fakeRemover) Confirm(*bool) (func(), error) {
	if f.err != nil {
		return nil, f.err
	}
	return func() { close(f.removed) }, nil
}

func TestAnAgreedRemovalStartsAndTheOtherAnswersAreReported(t *testing.T) {
	removed := make(chan struct{})
	outcome, err := Uninstall(fakeRemover{removed: removed})()
	if outcome != tui.RemovalStarted || err != nil {
		t.Fatalf("outcome %d, err %v", outcome, err)
	}
	select {
	case <-removed:
	case <-time.After(5 * time.Second):
		t.Fatal("the removal never ran")
	}

	if outcome, _ := Uninstall(fakeRemover{err: uninstall.ErrDeclined})(); outcome != tui.RemovalDeclined {
		t.Errorf("declined: outcome %d", outcome)
	}
	if outcome, _ := Uninstall(fakeRemover{err: uninstall.ErrNoDialog})(); outcome != tui.RemovalUnavailable {
		t.Errorf("no dialog: outcome %d", outcome)
	}
	if _, err := Uninstall(fakeRemover{err: errors.New("boom")})(); err == nil {
		t.Error("another failure should be reported")
	}
}
