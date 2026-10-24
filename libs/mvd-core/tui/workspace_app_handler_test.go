package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	tea "charm.land/bubbletea/v2"

	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/prefspopup"
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

func openedPreferences(t *testing.T) (a workspaceApp, x0, y0, w, h int) {
	t.Helper()
	a = NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: config.Config{OutputDir: "/music", MaxConcurrentDownloads: 2}, CfgPath: filepath.Join(t.TempDir(), "config.conf")},
	}).(workspaceApp)
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	next, _ = next.Update(workspace.PreferencesRequestedMsg{})
	a = next.(workspaceApp)
	w, h = a.popupSize()
	x0, y0 = dialog.Centre(a.width, a.height, w, h)
	return a, x0, y0, w, h
}

func TestAClickIsWorkedOutInsideTheFrameOfThePopup(t *testing.T) {
	a, x0, y0, _, _ := openedPreferences(t)
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
}

func TestAClickOnATabOfThePopupShowsItsPage(t *testing.T) {
	a, x0, y0, _, _ := openedPreferences(t)
	// The Download tab: the second box of the row, a few columns in.
	next, _ := a.Update(tea.MouseClickMsg{X: x0 + 1 + 14, Y: y0 + 2, Button: tea.MouseLeft})
	if got := next.(workspaceApp).prefs.Tab(); got != 1 {
		t.Errorf("tab %d, want 1", got)
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

func TestEscClosesThePopupAndSaveKeepsTheFolders(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	_, cmd := a.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	next, _ := a.Update(cmd())
	if next.(workspaceApp).inPrefs {
		t.Error("esc did not close the popup")
	}

	a, _, _, _, _ = openedPreferences(t)
	a.prefs = a.prefs.WithConfig(config.Config{OutputDir: "/elsewhere", MaxConcurrentDownloads: 2})
	next, _ = a.Update(prefspopup.SaveMsg{})
	if got := next.(workspaceApp).cfg.OutputDir; got != "/elsewhere" {
		t.Errorf("save left the folder at %q", got)
	}
}

func TestTheWorkspaceBehindThePreferencesOfTheFirstStartIsDrawn(t *testing.T) {
	a := NewWorkspaceApp(WorkspaceInput{
		Setup: SetupInput{Cfg: config.Config{OutputDir: "/music"}, CfgPath: filepath.Join(t.TempDir(), "config.conf"), OpenConfig: true},
	})
	next, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	screen := ansi.Strip(next.View().Content)
	if strings.Contains(screen, "starting...") || !strings.Contains(screen, "Preferences") || !strings.Contains(screen, "Add link") {
		t.Errorf("the first start shows\n%s", screen)
	}
}

func TestSavingPreferencesSaysSo(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	a.prefs = a.prefs.WithConfig(config.Config{OutputDir: "/elsewhere", MaxConcurrentDownloads: 2})
	next, _ := a.Update(prefspopup.SaveMsg{})
	screen := ansi.Strip(next.View().Content)
	if !strings.Contains(screen, "Preferences saved") {
		t.Errorf("saving says nothing:\n%s", screen)
	}
}

func TestLeavingPreferencesWithAChangedFolderWarnsThatItWasNotKept(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	a.prefs = a.prefs.WithConfig(config.Config{OutputDir: "/elsewhere", MaxConcurrentDownloads: 2})
	next, _ := a.Update(prefspopup.CloseMsg{})
	screen := ansi.Strip(next.View().Content)
	if !strings.Contains(screen, "were not kept") {
		t.Errorf("a dropped folder change is not warned of:\n%s", screen)
	}
}

func TestLeavingPreferencesWithNothingChangedSaysNothing(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	next, _ := a.Update(prefspopup.CloseMsg{})
	screen := ansi.Strip(next.View().Content)
	if strings.Contains(screen, "were not kept") || strings.Contains(screen, "Preferences saved") {
		t.Errorf("closing an untouched popup said something:\n%s", screen)
	}
}

func TestF1ShowsTheKeyboardShortcutsOverThePreferencesAndEscComesBackToThem(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	next, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyF1})
	a = next.(workspaceApp)
	if !a.shortcutsOpen || !strings.Contains(ansi.Strip(a.View().Content), "Keyboard shortcuts") {
		t.Fatalf("F1 did not open the shortcuts over the preferences")
	}
	next, _ = a.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	a = next.(workspaceApp)
	if a.shortcutsOpen || !a.inPrefs {
		t.Errorf("Esc should close the shortcuts and leave the preferences open: shortcuts %v, preferences %v", a.shortcutsOpen, a.inPrefs)
	}
	if !strings.Contains(ansi.Strip(a.View().Content), "Preferences") {
		t.Error("the preferences are not showing again")
	}
}

func TestClosingThePreferencesClosesTheShortcutsToo(t *testing.T) {
	a, _, _, _, _ := openedPreferences(t)
	a.shortcutsOpen = true
	if a = a.closePrefs(false); a.shortcutsOpen {
		t.Error("the shortcuts stayed open after the preferences closed")
	}
}
