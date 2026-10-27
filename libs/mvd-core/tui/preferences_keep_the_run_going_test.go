package tui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/workspace"
)

// firstMessages runs the commands an Init gives and returns the messages they bring, so a test can
// hand them back as the program would.
func firstMessages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var out []tea.Msg
	for _, c := range batch {
		out = append(out, firstMessages(c)...)
	}

	return out
}

// The download screen reads the run's events one at a time: each one handled asks for the next.
// One taken by the preferences instead was never handed on, and the screen stopped following the
// run for good: songs stayed at "finding" and counted as downloading after the preferences closed.
func TestTheRunsOwnMessagesReachTheScreenWhileThePreferencesAreOpen(t *testing.T) {
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: config.Config{OutputDir: "/music"}, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
	}).(workspaceApp)
	beat := firstMessages(a.Init())
	if len(beat) == 0 {
		t.Fatal("the workspace starts no beat to test with")
	}
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	next, _ = next.Update(workspace.PreferencesRequestedMsg{})
	if !next.(workspaceApp).inPrefs {
		t.Fatal("the preferences did not open")
	}
	for _, msg := range beat {
		if _, cmd := next.Update(msg); cmd == nil {
			t.Errorf("%T was swallowed by the preferences: the screen does not ask for the next one", msg)
		}
	}
}
