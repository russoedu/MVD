package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/settings"
)

func TestTheSetupScreensPassTheChooserToThePreferences(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{
		Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true,
		PickFolder: func(string) (string, bool, error) { return "/music", true, nil },
	})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m, cmd := m.Update(key("enter"))
	m, _ = m.Update(cmd())

	if got := m.(setupModel).config.Config().OutputDir; got != "/music" {
		t.Errorf("output dir %q", got)
	}
}

func TestTheMoveHintIsClickableOnThePreferences(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{
		Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true,
		Move: moveAnswering(MoveDeclined, nil),
	})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	for x := 0; x < 100; x++ {
		if h, ok := keybar.At(m.(setupModel).hints(), x); ok && h.Key == "m" {
			m, cmd := m.Update(click(x, 29))
			if m.(setupModel).config.Mode() != settings.EditConfirming || cmd == nil {
				t.Fatal("clicking the entry should ask")
			}
			return
		}
	}
	t.Fatal("the key bar should offer to move the app")
}

func TestTheUninstallHintIsClickableOnThePreferences(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{
		Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true,
		Uninstall: uninstallAnswering(RemovalDeclined, nil),
	})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	for x := 0; x < 100; x++ {
		if h, ok := keybar.At(m.(setupModel).hints(), x); ok && h.Key == "u" {
			m, cmd := m.Update(click(x, 29))
			if m.(setupModel).config.Mode() != settings.EditConfirming || cmd == nil {
				t.Fatal("clicking the entry should ask")
			}
			return
		}
	}
	t.Fatal("the key bar should offer uninstall")
}

func moveAnswering(outcome MoveOutcome, err error) Mover {
	return func() (MoveOutcome, error) { return outcome, err }
}

func uninstallAnswering(outcome RemovalOutcome, err error) Uninstaller {
	return func() (RemovalOutcome, error) { return outcome, err }
}
