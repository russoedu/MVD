package workspace

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestAConfirmationIsShownAtTheBottomWithATickAndGoesAtTheNextKey(t *testing.T) {
	m := listOf("https://a").Confirm("Preferences saved")
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "✓ Preferences saved") {
		t.Errorf("the confirmation is not shown:\n%s", screen)
	}
	m, _ = pressKey(m, tea.KeyPressMsg{Code: tea.KeyDown})
	if screen := ansi.Strip(m.render()); strings.Contains(screen, "Preferences saved") {
		t.Error("the confirmation should go away at the next key")
	}
}

func TestACautionIsShownAndAnErrorWinsOverIt(t *testing.T) {
	m := listOf("https://a").Caution("Not kept")
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "Not kept") {
		t.Errorf("the caution is not shown:\n%s", screen)
	}
	m.notice = "Cannot save the list: disk full"
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "Cannot save the list") || strings.Contains(screen, "Not kept") {
		t.Errorf("an error should be what is shown:\n%s", screen)
	}
}

func TestAConfirmationTakesAwayACaution(t *testing.T) {
	m := listOf("https://a").Caution("Not kept").Confirm("Saved")
	if m.caution != "" || m.done != "Saved" {
		t.Errorf("caution %q, done %q", m.caution, m.done)
	}
}
