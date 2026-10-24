package prefspopup

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/config"
)

func TestTheFirstStartWelcomesAndTheFirstKeyTakesTheWelcomeAway(t *testing.T) {
	m := New(config.Default("/app", "/music"), Host{}).SetSize(100, 30).WithWelcome()
	if !strings.Contains(ansi.Strip(m.View()), "Welcome! Check where your videos go") {
		t.Error("the first start does not welcome")
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if strings.Contains(ansi.Strip(m.View()), "Welcome!") {
		t.Error("the welcome should go at the first key")
	}
}

func TestPreferencesOpenedFromTheMenuDoNotWelcome(t *testing.T) {
	m := New(config.Default("/app", "/music"), Host{}).SetSize(100, 30)
	if strings.Contains(ansi.Strip(m.View()), "Welcome!") {
		t.Error("only the first start welcomes")
	}
}
