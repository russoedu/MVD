package tui

import (
	tea "charm.land/bubbletea/v2"

	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/workspace"
)

func TestPreferencesApplyAtOnceExceptWhereFilesGo(t *testing.T) {
	cfg := config.Config{OutputDir: "/music", MaxConcurrentDownloads: 2}
	var live engine.LiveSettings
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: cfg, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
		Live: func(c config.Config) engine.LiveSettings {
			live = engine.LiveSettings{Workers: c.MaxConcurrentDownloads}
			return live
		},
	}).(workspaceApp)

	changed := cfg
	changed.MaxConcurrentDownloads = 5
	changed.OutputDir = "/elsewhere"
	a = a.apply(changed, false)
	if a.cfg.MaxConcurrentDownloads != 5 || live.Workers != 5 {
		t.Errorf("the number of downloads was not applied: %d, %d", a.cfg.MaxConcurrentDownloads, live.Workers)
	}
	if a.cfg.OutputDir != "/music" {
		t.Errorf("the folder changed before it was confirmed: %q", a.cfg.OutputDir)
	}
	a = a.apply(changed, true)
	if a.cfg.OutputDir != "/elsewhere" {
		t.Errorf("the folder was not applied once confirmed: %q", a.cfg.OutputDir)
	}
}

func TestAClickOnThePopupKeyBarIsWorkedOutInsideItsFrame(t *testing.T) {
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: config.Config{OutputDir: "/music"}, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
	}).(workspaceApp)
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	a = next.(workspaceApp)
	next, _ = a.Update(workspace.PreferencesRequestedMsg{})
	a = next.(workspaceApp)

	w, h := a.popupSize()
	x0, y0 := dialog.Centre(a.width, a.height, w, h)
	click := func(x, y int) tea.MouseClickMsg { return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft} }

	if _, in := a.inPopup(click(x0, y0)); in {
		t.Errorf("a click on the frame reached the preferences")
	}
	if _, in := a.inPopup(click(0, 0)); in {
		t.Errorf("a click outside the popup reached the preferences")
	}
	got, in := a.inPopup(click(x0+3, y0+2))
	if !in || got.Mouse().X != 2 || got.Mouse().Y != 1 {
		t.Errorf("a click inside = %+v, %v", got, in)
	}
	// The last line of the inside is the key bar: one of its hints closes the popup.
	closed := false
	for x := 0; x < w-2 && !closed; x++ {
		_, cmd := a.Update(click(x0+1+x, y0+h-3))
		if cmd == nil {
			continue
		}
		if _, ok := cmd().(setupFinishedMsg); ok {
			closed = true
		}
	}
	if !closed {
		t.Errorf("no hint on the key bar of the popup closed it")
	}
}

func openedPreferences(t *testing.T) (workspaceApp, int, int, int, int) {
	t.Helper()
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: config.Config{OutputDir: "/music"}, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
	}).(workspaceApp)
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	next, _ = next.Update(workspace.PreferencesRequestedMsg{})
	a = next.(workspaceApp)
	w, h := a.popupSize()
	x0, y0 := dialog.Centre(a.width, a.height, w, h)
	return a, x0, y0, w, h
}

func TestTheTabsOfThePopupShowTheirScreens(t *testing.T) {
	a, x0, y0, _, _ := openedPreferences(t)
	_, spans := dialog.Tabs(popupTabs, 0)
	for tab, want := range map[int]int{tabAdvanced: screenAdvanced, tabColours: screenColours, tabPreferences: screenConfig} {
		click := tea.MouseClickMsg{X: x0 + 2 + spans[tab].X0, Y: y0 + 1, Button: tea.MouseLeft}
		next, _ := a.Update(click)
		a = next.(workspaceApp)
		if a.prefs.screen != want {
			t.Errorf("tab %d shows screen %d, want %d", tab, a.prefs.screen, want)
		}
	}
}

func TestTheButtonsOfThePopupCloseIt(t *testing.T) {
	_, spans := dialog.Buttons(popupButtons, -1, nil)
	for button, name := range popupButtons {
		a, x0, y0, _, h := openedPreferences(t)
		click := tea.MouseClickMsg{X: x0 + 2 + spans[button].X0, Y: y0 + h - 2, Button: tea.MouseLeft}
		next, _ := a.Update(click)
		if next.(workspaceApp).inPrefs {
			t.Errorf("%s did not close the popup", name)
		}
	}
}
