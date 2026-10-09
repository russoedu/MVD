package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

type outliner interface{ Outline() ScreenOutline }

func TestOutlineDescribesTheListScreen(t *testing.T) {
	m := NewSetupModel(SetupInput{URLs: []string{"https://a", "https://b"}})
	out := m.(outliner).Outline()

	if out.Title != "MVD · Download list (2 items)" || !out.HasText || out.Text != "https://a\nhttps://b" {
		t.Fatalf("unexpected outline %+v", out)
	}
	if len(out.Keys) != 5 || out.Keys[0] != (OutlineKey{Key: "ctrl+s", Description: "start"}) {
		t.Fatalf("unexpected keys %+v", out.Keys)
	}
}

func TestOutlineFollowsTheConfigScreenSelection(t *testing.T) {
	m, _ := NewSetupModel(SetupInput{Cfg: config.Config{VideoQuality: "best"}, OpenConfig: true}).Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	out := m.(outliner).Outline()

	if out.Title != "MVD · Preferences" || len(out.Items) != cfgItemCount {
		t.Fatalf("unexpected outline %+v", out)
	}
	if !out.Items[1].Selected || out.Items[1].Label != "Video Quality" || out.Items[1].Value != "best" {
		t.Fatalf("unexpected selection %+v", out.Items[1])
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // open the radio editor
	editing := m.(outliner).Outline().Items[1]
	if !editing.Editing || len(editing.Choices) == 0 {
		t.Fatalf("expected the radio editor to be described, got %+v", editing)
	}
}
