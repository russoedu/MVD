package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
)

func preferencesWith(uninstall Uninstaller) configModel {
	return newConfigModel(config.Default("/app", "/dl")).withUninstaller(uninstall).setSize(100, 30)
}

func uninstallAnswering(outcome RemovalOutcome, err error) Uninstaller {
	return func() (RemovalOutcome, error) { return outcome, err }
}

func TestTheUninstallKeyAsksAndSaysHowItWent(t *testing.T) {
	cases := []struct {
		name    string
		outcome RemovalOutcome
		err     error
		want    string
	}{
		{"agreed", RemovalStarted, nil, "Removing MVD. This window will close."},
		{"declined", RemovalDeclined, nil, "Nothing was removed."},
		{"no way to ask", RemovalUnavailable, nil, "This machine cannot ask for confirmation, so nothing was removed. Start MVD with -uninstall from a terminal."},
		{"failed", RemovalDeclined, errors.New("boom"), "The removal could not be asked: boom"},
	}
	for _, c := range cases {
		m := preferencesWith(uninstallAnswering(c.outcome, c.err))

		m, cmd, _, _ := m.update(key("u"))
		if m.mode != editConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.mode, cmd)
		}
		m, _, _, _ = m.update(cmd())
		if m.mode != editNone || m.status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.mode, m.status, c.want)
		}
		if got := m.outline().Prompt; got != c.want {
			t.Errorf("%s: the outline should say it too, got %q", c.name, got)
		}
		if view := m.view(140, 30); !strings.Contains(view, c.want[:20]) {
			t.Errorf("%s: the screen should show the status", c.name)
		}
	}
}

func TestWhileTheQuestionsAreOpenKeysAreIgnored(t *testing.T) {
	m := preferencesWith(uninstallAnswering(RemovalDeclined, nil))
	m, _, _, _ = m.update(key("u"))

	m, _, out, _ := m.update(key("s"))
	if m.mode != editConfirming || out != cfgNone {
		t.Errorf("mode %d, outcome %d: keys should wait for the answers", m.mode, out)
	}
	if m.outline().Prompt == "" || len(m.hints()) != 0 {
		t.Errorf("prompt %q, hints %v", m.outline().Prompt, m.hints())
	}
}

func TestAnyKeyClearsTheStatus(t *testing.T) {
	m := preferencesWith(uninstallAnswering(RemovalDeclined, nil))
	m, cmd, _, _ := m.update(key("u"))
	m, _, _, _ = m.update(cmd())
	m, _, _, _ = m.update(key("down"))

	if m.status != "" {
		t.Errorf("status %q", m.status)
	}
}

func TestWithoutAnUninstallerThereIsNoKeyAndNoHint(t *testing.T) {
	m := preferencesWith(nil)

	m, cmd, _, _ := m.update(key("u"))
	if m.mode != editNone || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.mode, cmd)
	}
	for _, h := range m.hints() {
		if h.key == "u" {
			t.Error("the key bar should not offer uninstall")
		}
	}
}

func TestTheUninstallHintIsClickableOnThePreferences(t *testing.T) {
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{
		Cfg: config.Default(dir, dir), CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt", OpenConfig: true,
		Uninstall: uninstallAnswering(RemovalDeclined, nil),
	})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	for x := 0; x < 100; x++ {
		if h, ok := hintAt(m.(setupModel).hints(), x); ok && h.key == "u" {
			m, cmd := m.Update(click(x, 29))
			if m.(setupModel).config.mode != editConfirming || cmd == nil {
				t.Fatal("clicking the entry should ask")
			}
			return
		}
	}
	t.Fatal("the key bar should offer uninstall")
}
