package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
)

func TestInANarrowWindowTheBottomLineSaysTheButtonsAreInTheRunMenu(t *testing.T) {
	m := New(Host{}, config.Config{}, downloadlist.New([]string{"https://a"}))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	if screen := screenOf(next.(Model)); !strings.Contains(screen, "use the Run menu") {
		t.Errorf("a narrow window does not say where the buttons went: %s", screen)
	}
}

func TestInANarrowWindowAFinishedRunStillSaysTheButtonsAreInTheRunMenu(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 0)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	if screen := screenOf(next.(Model)); !strings.Contains(screen, "Run menu") {
		t.Errorf("a finished run in a narrow window does not say where the buttons went: %s", screen)
	}
}

func TestInAWideWindowTheButtonsAreThereAndNothingSaysOtherwise(t *testing.T) {
	if screen := screenOf(listOf("https://a")); strings.Contains(screen, "use the Run menu") {
		t.Errorf("a wide window talks about the Run menu: %s", screen)
	}
}
