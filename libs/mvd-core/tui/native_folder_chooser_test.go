package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
)

// chooseOnFirstSetting presses enter on the Output Folder setting with the given
// chooser and returns the model, the command it started, and the folder the chooser was asked to start at.
func chooseOnFirstSetting(t *testing.T, pick FolderPicker) (configModel, tea.Cmd) {
	t.Helper()
	m := newConfigModel(config.Default("/app", "/dl")).withFolderPicker(pick).setSize(100, 30)
	next, cmd, _, _ := m.update(key("enter"))
	return next, cmd
}

func TestAFolderSettingOpensTheHostsChooserAndTakesItsAnswer(t *testing.T) {
	var startedAt string
	m, cmd := chooseOnFirstSetting(t, func(start string) (string, bool, error) {
		startedAt = start
		return "/music", true, nil
	})

	if m.mode != editPicking || cmd == nil {
		t.Fatalf("mode %d, cmd %v: the chooser should be open", m.mode, cmd)
	}
	answer := cmd()
	if startedAt != "/dl" {
		t.Errorf("the chooser should start at the current folder, started at %q", startedAt)
	}

	m, _, _, _ = m.update(answer)
	if m.mode != editNone || m.cfg.OutputDir != "/music" {
		t.Errorf("mode %d, output dir %q", m.mode, m.cfg.OutputDir)
	}
}

func TestCancellingTheChooserChangesNothing(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })
	m, _, _, _ = m.update(cmd())

	if m.mode != editNone || m.cfg.OutputDir != "/dl" {
		t.Errorf("mode %d, output dir %q", m.mode, m.cfg.OutputDir)
	}
}

func TestWithoutAChooserOnTheMachineTheBuiltInBrowserOpens(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, errors.New("no desktop") })
	m, _, _, _ = m.update(cmd())

	if m.mode != editFolder {
		t.Errorf("mode %d, want the built-in folder browser", m.mode)
	}
}

func TestWithoutAHostChooserTheBuiltInBrowserOpensAtOnce(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, nil)

	if m.mode != editFolder || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.mode, cmd)
	}
}

func TestTheLogFolderUsesTheChooserToo(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl")).withFolderPicker(func(start string) (string, bool, error) {
		return "/logs", true, nil
	}).setSize(100, 30)
	m.cursor = 9
	m, cmd, _, _ := m.update(key("enter"))
	m, _, _, _ = m.update(cmd())

	if m.cfg.LogDir != "/logs" {
		t.Errorf("log dir %q", m.cfg.LogDir)
	}
}

func TestWhileTheChooserIsOpenKeysAreIgnoredAndTheScreenSaysWhy(t *testing.T) {
	m, _ := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })

	m, _, _, _ = m.update(key("esc"))
	m, _, out, _ := m.update(key("s"))
	if m.mode != editPicking || out != cfgNone {
		t.Errorf("mode %d, outcome %d: keys should wait for the chooser", m.mode, out)
	}
	if got := m.outline().Prompt; got != "Choose a folder in the window that opened" {
		t.Errorf("prompt %q", got)
	}
	if len(m.hints()) != 0 {
		t.Errorf("no key hints while waiting, got %v", m.hints())
	}
}

func TestTheSetupScreensPassTheChooserToThePreferences(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{
		Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true,
		PickFolder: func(string) (string, bool, error) { return "/music", true, nil },
	})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m, cmd := m.Update(key("enter"))
	m, _ = m.Update(cmd())

	if got := m.(setupModel).config.cfg.OutputDir; got != "/music" {
		t.Errorf("output dir %q", got)
	}
}
