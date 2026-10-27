package tui

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/workspace"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
)

// styledApp is the workspace app of a host that draws the classic styles, with what it was told.
func styledApp(t *testing.T, style string) (workspaceApp, *[]string, *[]windowsnapshot.Snapshot) {
	t.Helper()
	styles, windows := &[]string{}, &[]windowsnapshot.Snapshot{}
	a := NewWorkspaceApp(WorkspaceInput{
		Setup:  SetupInput{Cfg: config.Config{OutputDir: "/music", WindowStyle: style}, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
		Style:  func(s string) { *styles = append(*styles, s) },
		Window: func(s windowsnapshot.Snapshot) { *windows = append(*windows, s) },
	}).(workspaceApp)
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	return next.(workspaceApp), styles, windows
}

func TestTheTerminalLookIsReportedAndDrawsNoWindow(t *testing.T) {
	_, styles, windows := styledApp(t, "")
	if len(*styles) != 1 || (*styles)[0] != config.WindowStyleDOS || len(*windows) != 0 {
		t.Errorf("styles %v, %d windows", *styles, len(*windows))
	}
}

func TestAClassicStyleGetsWhatTheWindowShowsAndThePreferencesCoverIt(t *testing.T) {
	a, styles, windows := styledApp(t, "mac-os-9")
	if len(*styles) != 1 || (*styles)[0] != "mac-os-9" || len(*windows) != 1 || (*windows)[0].Covered != nil {
		t.Fatalf("styles %v, windows %+v", *styles, *windows)
	}
	next, _ := a.Update(workspace.PreferencesRequestedMsg{})
	last := (*windows)[len(*windows)-1]
	if last.Covered == nil || last.Covered.Title != "Preferences" {
		t.Errorf("the preferences do not cover the window: %+v", last.Covered)
	}

	// Another style chosen reaches the host at the next change.
	app := next.(workspaceApp).closePrefs(false)
	changed := app.cfg
	changed.WindowStyle = config.WindowStyleDOS
	app.apply(changed, false).Update(tea.KeyPressMsg{})
	if (*styles)[len(*styles)-1] != config.WindowStyleDOS {
		t.Errorf("styles %v", *styles)
	}
}
