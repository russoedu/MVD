package colours

import (
	"testing"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/testing/keys"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func TestColoursModelEditsAndValidates(t *testing.T) {
	t.Cleanup(func() { look.Apply(config.DefaultThemeColors()) })
	m := New(config.Default("/app", "/dl"))

	// Edit the accent: a bad value is refused and keeps the editor open.
	m, _, _, _ = m.Update(keys.Press("enter"))
	m.input.SetValue("purple")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if !m.editing || m.status == "" || m.cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Fatalf("an invalid colour must be refused: editing=%v status=%q accent=%q", m.editing, m.status, m.cfg.Colors.Accent)
	}

	// A valid one is accepted and applied to the interface.
	m.input.SetValue("#123456")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.editing || m.cfg.Colors.Accent != "#123456" || look.Title.GetForeground() != lipgloss.Color("#123456") {
		t.Errorf("a valid colour should be accepted and applied: %+v", m.cfg.Colors)
	}

	// d puts the default back.
	m, _, _, _ = m.Update(keys.Press("d"))
	if m.cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Errorf("d should restore the default, got %q", m.cfg.Colors.Accent)
	}

	// Moving down and editing changes another slot, and s saves.
	m, _, _, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _, _, _ = m.Update(keys.Press("enter"))
	m.input.SetValue("#abc")
	m, _, _, _ = m.Update(keys.Press("enter"))
	_, _, out, cfg := m.Update(keys.Press("s"))
	if out != Save || cfg.Colors.Focus != "#abc" {
		t.Errorf("s should save the edited colours, got out=%d focus=%q", out, cfg.Colors.Focus)
	}
}
