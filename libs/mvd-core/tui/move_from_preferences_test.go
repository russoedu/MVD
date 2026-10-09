package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func preferencesThatMove(move Mover) configModel {
	return newConfigModel(config.Default("/app", "/dl")).withMover(move).setSize(100, 30)
}

func moveAnswering(outcome MoveOutcome, err error) Mover {
	return func() (MoveOutcome, error) { return outcome, err }
}

func TestTheMoveKeyAsksAndSaysHowItWent(t *testing.T) {
	cases := []struct {
		name    string
		outcome MoveOutcome
		err     error
		want    string
	}{
		{"agreed", MoveStarted, nil, "MVD was moved to its own folder. This window will close and the new copy will open."},
		{"declined", MoveDeclined, nil, "MVD was not moved."},
		{"already there", MoveAlreadyThere, nil, "MVD already lives in its own folder."},
		{"nowhere to go", MoveUnavailable, nil, "This machine has no place to move MVD to, or no way to ask. Start MVD with -move from a terminal."},
		{"failed", MoveDeclined, errors.New("boom"), "MVD could not be moved: boom"},
	}
	for _, c := range cases {
		m := preferencesThatMove(moveAnswering(c.outcome, c.err))

		m, cmd, _, _ := m.update(key("m"))
		if m.mode != editConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.mode, cmd)
		}
		m, _, _, _ = m.update(cmd())
		if m.mode != editNone || m.status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.mode, m.status, c.want)
		}
		if view := m.view(140, 30); !strings.Contains(view, c.want[:min(len(c.want), 15)]) {
			t.Errorf("%s: the screen should show the status", c.name)
		}
	}
}

func TestWithoutAMoverThereIsNoMoveKeyAndNoHint(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl")).setSize(100, 30)

	m, cmd, _, _ := m.update(key("m"))
	if cmd != nil || m.mode != editNone {
		t.Errorf("mode %d, cmd %v: nothing should happen", m.mode, cmd)
	}
	for _, h := range m.hints() {
		if h.key == "m" {
			t.Error("the key bar should not offer to move")
		}
	}
}

func TestTheMoveHintIsOfferedBesideUninstall(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl")).
		withMover(moveAnswering(MoveDeclined, nil)).
		withUninstaller(uninstallAnswering(RemovalDeclined, nil)).
		setSize(100, 30)

	var keys []string
	for _, h := range m.hints() {
		keys = append(keys, h.key)
	}
	got := strings.Join(keys, " ")
	if !strings.Contains(got, "m u") {
		t.Errorf("hints %q, want move and uninstall side by side", got)
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
		if h, ok := hintAt(m.(setupModel).hints(), x); ok && h.key == "m" {
			m, cmd := m.Update(click(x, 29))
			if m.(setupModel).config.mode != editConfirming || cmd == nil {
				t.Fatal("clicking the entry should ask")
			}
			return
		}
	}
	t.Fatal("the key bar should offer to move the app")
}
