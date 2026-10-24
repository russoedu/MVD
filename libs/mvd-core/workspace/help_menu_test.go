package workspace

import (
	"net/url"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/menubar"
)

func TestEveryRunEntryHasItsOwnShortcutAndOnlyPauseAndResumeShareOne(t *testing.T) {
	seen := map[string]string{}
	for _, mn := range menus() {
		if mn.Name != "Run" {
			continue
		}
		for _, e := range mn.Entries {
			if e.Separator {
				continue
			}
			if e.Shortcut == "" {
				t.Errorf("%q has no shortcut", e.Name)
			}
			if other, ok := seen[e.Shortcut]; ok && (e.Name != entryResume || other != entryPause) {
				t.Errorf("%q and %q share %s", other, e.Name, e.Shortcut)
			}
			seen[e.Shortcut] = e.Name
		}
	}
}

func TestNoShortcutIsGivenToTwoDifferentThingsAnywhere(t *testing.T) {
	seen := map[string]string{}
	for _, mn := range menus() {
		for _, e := range mn.Entries {
			if e.Separator || e.Shortcut == "" {
				continue
			}
			if other, ok := seen[e.Shortcut]; ok && (e.Name != entryResume || other != entryPause) {
				t.Errorf("%q and %q share %s", other, e.Name, e.Shortcut)
			}
			seen[e.Shortcut] = e.Name
		}
	}
}

func TestTheShortcutsWindowListsEveryShortcutOfTheMenusAndNothingElseOfTheirs(t *testing.T) {
	m := listOf("https://a")
	next, _ := m.Update(menubar.SelectedMsg{Name: entryShortcuts})
	m = next.(Model)
	if !m.shortcuts || !m.WindowOpen() {
		t.Fatal("the Keyboard shortcuts window did not open")
	}
	all := ansi.Strip(strings.Join(shortcutLines(), "\n"))
	for _, mn := range menus() {
		for _, e := range mn.Entries {
			if e.Separator || e.Shortcut == "" {
				continue
			}
			if !strings.Contains(all, e.Name) || !strings.Contains(all, menubar.ShortcutLabel(e.Shortcut)) {
				t.Errorf("the window lacks %q or its shortcut %s", e.Name, e.Shortcut)
			}
		}
	}
	screen := ansi.Strip(m.render())
	for _, want := range []string{"Keyboard shortcuts", "File", "esc close"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the window lacks %q", want)
		}
	}
	if !strings.Contains(all, "While downloading") {
		t.Error("the keys of the download screen are missing")
	}
	if !strings.Contains(all, "Pause / Resume") {
		t.Error("Pause and Resume share Ctrl+P and should be said once")
	}
}

func TestF1OpensTheShortcutsWindowAndEscClosesIt(t *testing.T) {
	m := listOf("https://a")
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyF1})
	if !m.shortcuts {
		t.Fatal("F1 should open the Keyboard shortcuts window")
	}
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyDown})
	if m.shortcutsScroll == 0 && m.shortcutsLayout().maxScroll > 0 {
		t.Error("down should scroll a window that is taller than the screen")
	}
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.shortcuts {
		t.Error("esc should close it")
	}
}

func TestTheShortcutsWindowStillShowsItsFooterOnASmallScreen(t *testing.T) {
	m := New(Host{}, config.Config{}, downloadlist.New([]string{"a"}))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	m = next.(Model)
	m.shortcuts = true
	screen := ansi.Strip(m.render())
	if !strings.Contains(screen, "↑↓ scroll  esc close") {
		t.Errorf("the footer is missing on a small screen:\n%s", screen)
	}
}

func TestDocumentationAndReportAProblemOpenTheBrowser(t *testing.T) {
	var opened []string
	m := New(Host{Version: "1.2.3", Open: func(u string) error { opened = append(opened, u); return nil }}, config.Config{}, downloadlist.New([]string{"a"}))
	m = sized(m)
	next, _ := m.Update(menubar.SelectedMsg{Name: entryDocs})
	_, _ = next.(Model).Update(menubar.SelectedMsg{Name: entryReport})
	if len(opened) != 2 {
		t.Fatalf("opened %v", opened)
	}
	if !strings.HasPrefix(opened[0], projectURL) {
		t.Errorf("the documentation opened %q", opened[0])
	}
	u, err := url.Parse(opened[1])
	if err != nil || !strings.HasSuffix(u.Path, "/issues/new") {
		t.Fatalf("the report opened %q (%v)", opened[1], err)
	}
	if body := u.Query().Get("body"); !strings.Contains(body, "MVD 1.2.3") {
		t.Errorf("the report does not carry the version: %q", body)
	}
}

func TestWithoutABrowserTheAddressIsSaidInstead(t *testing.T) {
	m := listOf("https://a")
	next, _ := m.Update(menubar.SelectedMsg{Name: entryDocs})
	if got := next.(Model).notice; !strings.Contains(got, projectURL) {
		t.Errorf("the notice is %q", got)
	}
}

// pressKey gives a key press to the model.
func pressKey(m Model, key tea.KeyPressMsg) (Model, tea.Cmd) {
	next, cmd := m.Update(key)
	m = next.(Model)
	// A shortcut of the menu answers with the entry that was chosen; the program feeds it back.
	if cmd != nil {
		if selected, ok := cmd().(menubar.SelectedMsg); ok {
			next, cmd = m.Update(selected)
			return next.(Model), cmd
		}
	}
	return m, cmd
}
