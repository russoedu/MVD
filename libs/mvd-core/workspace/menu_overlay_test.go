package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
)

func TestAnOpenMenuCoversOnlyItsOwnBox(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	closed := strings.Split(ansi.Strip(m.render()), "\n")
	m.bar.SetOpen(0)
	open := strings.Split(ansi.Strip(m.render()), "\n")
	if len(open) != len(closed) {
		t.Fatalf("%d rows, was %d", len(open), len(closed))
	}
	if !strings.Contains(strings.Join(open, "\n"), entryQuit) {
		t.Error("the menu is not shown")
	}
	row := 3 // inside the box
	if !strings.HasSuffix(open[row], closed[row][len(closed[row])-30:]) {
		t.Errorf("the screen right of the menu was covered: %q", open[row])
	}
}

func TestAnOpenMenuFollowsTheMouseAndMovingOverAnotherHeadingOpensIt(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	if m.View().MouseMode == tea.MouseModeAllMotion {
		t.Fatal("the mouse is followed with no menu open")
	}
	m.bar.SetOpen(0)
	if m.View().MouseMode != tea.MouseModeAllMotion {
		t.Fatal("the mouse is not followed while a menu is open")
	}
	// The Edit heading is on the bar row, after MVD, a rule and File.
	x := strings.Index(ansi.Strip(m.barRow()), "Edit") + 1
	next, _ := m.Update(tea.MouseMotionMsg{X: x, Y: 0})
	if got := next.(Model).bar.Open(); got != 1 {
		t.Errorf("open menu %d, want 1 (Edit)", got)
	}
}
