package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
)

func TestConfigColoursTransition(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	if _, _, out, _ := m.update(key("c")); out != cfgColours {
		t.Errorf("'c' should request the colours screen, got %d", out)
	}
}

func TestColoursModelEditsAndValidates(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.DefaultThemeColors()) })
	m := newColoursModel(config.Default("/app", "/dl"))

	// Edit the accent: a bad value is refused and keeps the editor open.
	m, _, _, _ = m.update(key("enter"))
	m.input.SetValue("purple")
	m, _, _, _ = m.update(key("enter"))
	if !m.editing || m.status == "" || m.cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Fatalf("an invalid colour must be refused: editing=%v status=%q accent=%q", m.editing, m.status, m.cfg.Colors.Accent)
	}

	// A valid one is accepted and applied to the interface.
	m.input.SetValue("#123456")
	m, _, _, _ = m.update(key("enter"))
	if m.editing || m.cfg.Colors.Accent != "#123456" || string(colMagenta) != "#123456" {
		t.Errorf("a valid colour should be accepted and applied: %+v", m.cfg.Colors)
	}

	// d puts the default back.
	m, _, _, _ = m.update(key("d"))
	if m.cfg.Colors.Accent != config.DefaultThemeColors().Accent {
		t.Errorf("d should restore the default, got %q", m.cfg.Colors.Accent)
	}

	// Moving down and editing changes another slot, and s saves.
	m, _, _, _ = m.update(tea.KeyMsg{Type: tea.KeyDown})
	m, _, _, _ = m.update(key("enter"))
	m.input.SetValue("#abc")
	m, _, _, _ = m.update(key("enter"))
	_, _, out, cfg := m.update(key("s"))
	if out != colSave || cfg.Colors.Focus != "#abc" {
		t.Errorf("s should save the edited colours, got out=%d focus=%q", out, cfg.Colors.Focus)
	}
}
