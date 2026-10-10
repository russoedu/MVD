package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func click(x, y int) tea.MouseMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func wheel(x, y int, button tea.MouseButton) tea.MouseMsg {
	return tea.MouseWheelMsg{X: x, Y: y, Button: button}
}

func TestHintAtFindsTheEntryDrawnUnderAColumn(t *testing.T) {
	hints := []keyHint{{"ctrl+s", "start"}, {"ctrl+p", "preferences"}}

	// " ctrl+s start  ctrl+p preferences": the first entry fills columns 1-12, the second 15-32.
	cases := []struct {
		x    int
		want string
	}{{1, "ctrl+s"}, {12, "ctrl+s"}, {15, "ctrl+p"}, {32, "ctrl+p"}}
	for _, c := range cases {
		if got, ok := hintAt(hints, c.x); !ok || got.key != c.want {
			t.Errorf("column %d: got %q, %v, want %q", c.x, got.key, ok, c.want)
		}
	}
	for _, x := range []int{0, 13, 14, 33} {
		if _, ok := hintAt(hints, x); ok {
			t.Errorf("column %d is between entries, not on one", x)
		}
	}
}

func TestSetupScreensKeyBarSettingsAndWheel(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{Cfg: config.Default(dir, dir), URLs: []string{"https://a"}, CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt"})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Clicking "preferences" in the key bar opens the preferences.
	m, _ = m.Update(click(48, 29))
	if m.(setupModel).screen != screenConfig {
		t.Fatalf("the click should open the preferences, on screen %d", m.(setupModel).screen)
	}

	// Clicking a setting selects it; the title is row 0, so the third setting is row 3.
	m, _ = m.Update(click(10, 3))
	if got := m.(setupModel).config.cursor; got != 2 {
		t.Errorf("want the third setting selected, got %d", got)
	}

	// The wheel moves like the arrow keys.
	m, _ = m.Update(wheel(10, 5, tea.MouseWheelDown))
	if got := m.(setupModel).config.cursor; got != 5 {
		t.Errorf("a notch down moves three settings, got cursor %d", got)
	}
	m, _ = m.Update(wheel(10, 5, tea.MouseWheelUp))
	if got := m.(setupModel).config.cursor; got != 2 {
		t.Errorf("a notch up moves three back, got cursor %d", got)
	}

	// A click on the bar, away from any entry, does nothing; clicking "cancel" returns to the list.
	m, _ = m.Update(click(0, 29))
	if m.(setupModel).screen != screenConfig {
		t.Error("a click between entries should do nothing")
	}
	hints := m.(setupModel).hints()
	cancelAt := 1
	for _, h := range hints[:len(hints)-1] {
		cancelAt += len(h.key) + 1 + len(h.desc) + 2
	}
	m, _ = m.Update(click(cancelAt, 29))
	if m.(setupModel).screen != screenList {
		t.Errorf("clicking cancel should return to the list, on screen %d", m.(setupModel).screen)
	}
}

func TestSetupScreensIgnoreClicksOnSettingsWhileOneIsBeingEdited(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("enter")) // opens an editor on the second setting

	before := m.(setupModel).config.cursor
	m, _ = m.Update(click(10, 6))
	if got := m.(setupModel).config.cursor; got != before {
		t.Errorf("a click while editing moved the selection from %d to %d", before, got)
	}
}

func TestDownloadScreenClicksSelectPanesRowsAndKeyBarEntries(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	// Wide layout: the playlists box fills rows 3-7 (its rows 4-6), the entries box starts at row 8
	// (its rows 9 and below), and the output box starts at column 45.

	next, _ := m.Update(click(5, 5))
	m = next.(model)
	if m.selPlaylist != 1 || m.focus != paneLists {
		t.Errorf("clicking the second playlist: selPlaylist %d, focus %d", m.selPlaylist, m.focus)
	}

	m.selPlaylist = 0
	next, _ = m.Update(click(5, 11))
	m = next.(model)
	if m.selEntry != 2 || m.focus != paneEntries || m.follow {
		t.Errorf("clicking the third entry: selEntry %d, focus %d, follow %v", m.selEntry, m.focus, m.follow)
	}

	m.focus = paneLists
	next, _ = m.Update(click(11, 35)) // "tab pane" in the key bar
	m = next.(model)
	if m.focus != paneEntries {
		t.Errorf("clicking tab in the key bar should change pane, focus %d", m.focus)
	}
}

func TestDownloadScreenWheelScrollsThePaneUnderThePointer(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.selEntry = 4 // the entry with a long log

	next, _ := m.Update(wheel(80, 10, tea.MouseWheelUp)) // over the output box
	m = next.(model)
	if m.focus != paneOutput || m.outScroll != wheelStep {
		t.Errorf("wheel up over the output: focus %d, scrolled %d", m.focus, m.outScroll)
	}

	next, _ = m.Update(wheel(5, 11, tea.MouseWheelDown)) // over the entries
	m = next.(model)
	if m.focus != paneEntries || m.selEntry != 5 {
		t.Errorf("wheel down over the entries: focus %d, selEntry %d", m.focus, m.selEntry)
	}
}

func TestDownloadScreenIgnoresTheMouseWhileConfirmingToQuit(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.confirmQuit = true

	next, _ := m.Update(click(5, 5))
	if next.(model).selPlaylist != m.selPlaylist {
		t.Error("a click while confirming should do nothing")
	}
}
