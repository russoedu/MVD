package terminalui

import (
	"errors"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/tui"
	"youtube-downloader/libs/mvd-server/api"
)

type fakeUninstaller struct {
	err     error
	removed chan struct{}
}

func (f fakeUninstaller) Confirm(*bool) (func(), error) {
	if f.err != nil {
		return nil, f.err
	}
	return func() { close(f.removed) }, nil
}

func TestAnAgreedRemovalStartsAndTheOtherAnswersAreReported(t *testing.T) {
	removed := make(chan struct{})
	outcome, err := Uninstall(fakeUninstaller{removed: removed})()
	if outcome != tui.RemovalStarted || err != nil {
		t.Fatalf("outcome %d, err %v", outcome, err)
	}
	select {
	case <-removed:
	case <-time.After(5 * time.Second):
		t.Fatal("the removal never ran")
	}

	if outcome, _ := Uninstall(fakeUninstaller{err: api.ErrUninstallDeclined})(); outcome != tui.RemovalDeclined {
		t.Errorf("declined: outcome %d", outcome)
	}
	if outcome, _ := Uninstall(fakeUninstaller{err: api.ErrNoDialog})(); outcome != tui.RemovalUnavailable {
		t.Errorf("no dialog: outcome %d", outcome)
	}
	if _, err := Uninstall(fakeUninstaller{err: errors.New("boom")})(); err == nil {
		t.Error("another failure should be reported")
	}
}
