package tui

import (
	"testing"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/testing/mouse"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func click(x, y int) tea.MouseMsg { return mouse.Click(x, y) }

func wheel(x, y int, button tea.MouseButton) tea.MouseMsg { return mouse.Wheel(x, y, button) }

func TestHintAtFindsTheEntryDrawnUnderAColumn(t *testing.T) {
	hints := []keybar.Hint{{Key: "ctrl+s", Desc: "start"}, {Key: "ctrl+p", Desc: "preferences"}}

	// " ctrl+s start  ctrl+p preferences": the first entry fills columns 1-12, the second 15-32.
	cases := []struct {
		x    int
		want string
	}{{1, "ctrl+s"}, {12, "ctrl+s"}, {15, "ctrl+p"}, {32, "ctrl+p"}}
	for _, c := range cases {
		if got, ok := keybar.At(hints, c.x); !ok || got.Key != c.want {
			t.Errorf("column %d: got %q, %v, want %q", c.x, got.Key, ok, c.want)
		}
	}
	for _, x := range []int{0, 13, 14, 33} {
		if _, ok := keybar.At(hints, x); ok {
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
	if got := m.(setupModel).config.Cursor(); got != 2 {
		t.Errorf("want the third setting selected, got %d", got)
	}

	// The wheel moves like the arrow keys.
	m, _ = m.Update(wheel(10, 5, tea.MouseWheelDown))
	if got := m.(setupModel).config.Cursor(); got != 5 {
		t.Errorf("a notch down moves three settings, got cursor %d", got)
	}
	m, _ = m.Update(wheel(10, 5, tea.MouseWheelUp))
	if got := m.(setupModel).config.Cursor(); got != 2 {
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
		cancelAt += len(h.Key) + 1 + len(h.Desc) + 2
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

	before := m.(setupModel).config.Cursor()
	m, _ = m.Update(click(10, 6))
	if got := m.(setupModel).config.Cursor(); got != before {
		t.Errorf("a click while editing moved the selection from %d to %d", before, got)
	}
}
