package preferences

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"
	"youtube-downloader/libs/mvd-core/testing/keys"
)

// chooseOnFirstSetting presses enter on the Output Folder setting with the given

// chooser and returns the model, the command it started, and the folder the chooser was asked to start at.

func chooseOnFirstSetting(t *testing.T, pick FolderPicker) (Model, tea.Cmd) {
	t.Helper()
	m := New(config.Default("/app", "/dl")).WithFolderPicker(pick).SetSize(100, 30)
	next, cmd, _, _ := m.Update(keys.Press("enter"))
	return next, cmd
}

func TestAFolderSettingOpensTheHostsChooserAndTakesItsAnswer(t *testing.T) {
	var startedAt string
	m, cmd := chooseOnFirstSetting(t, func(start string) (string, bool, error) {
		startedAt = start
		return "/music", true, nil
	})

	if m.session.Mode != settings.EditPicking || cmd == nil {
		t.Fatalf("mode %d, cmd %v: the chooser should be open", m.session.Mode, cmd)
	}
	answer := cmd()
	if startedAt != "/dl" {
		t.Errorf("the chooser should start at the current folder, started at %q", startedAt)
	}

	m, _, _, _ = m.Update(answer)
	if m.session.Mode != settings.EditNone || m.session.Cfg.OutputDir != "/music" {
		t.Errorf("mode %d, output dir %q", m.session.Mode, m.session.Cfg.OutputDir)
	}
}

func TestCancellingTheChooserChangesNothing(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })
	m, _, _, _ = m.Update(cmd())

	if m.session.Mode != settings.EditNone || m.session.Cfg.OutputDir != "/dl" {
		t.Errorf("mode %d, output dir %q", m.session.Mode, m.session.Cfg.OutputDir)
	}
}

func TestWithoutAChooserOnTheMachineTheBuiltInBrowserOpens(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, errors.New("no desktop") })
	m, _, _, _ = m.Update(cmd())

	if m.session.Mode != settings.EditFolder {
		t.Errorf("mode %d, want the built-in folder browser", m.session.Mode)
	}
}

func TestWithoutAHostChooserTheBuiltInBrowserOpensAtOnce(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, nil)

	if m.session.Mode != settings.EditFolder || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.session.Mode, cmd)
	}
}

func TestTheLogFolderUsesTheChooserToo(t *testing.T) {
	m := New(config.Default("/app", "/dl")).WithFolderPicker(func(start string) (string, bool, error) {
		return "/logs", true, nil
	}).SetSize(100, 30)
	m.session.Cursor = 9
	m, cmd, _, _ := m.Update(keys.Press("enter"))
	m, _, _, _ = m.Update(cmd())

	if m.session.Cfg.LogDir != "/logs" {
		t.Errorf("log dir %q", m.session.Cfg.LogDir)
	}
}

func TestWhileTheChooserIsOpenKeysAreIgnoredAndTheScreenSaysWhy(t *testing.T) {
	m, _ := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })

	m, _, _, _ = m.Update(keys.Press("esc"))
	m, _, out, _ := m.Update(keys.Press("s"))
	if m.session.Mode != settings.EditPicking || out != None {
		t.Errorf("mode %d, outcome %d: keys should wait for the chooser", m.session.Mode, out)
	}
	if got := m.Outline().Prompt; got != "Choose a folder in the window that opened" {
		t.Errorf("prompt %q", got)
	}
	if len(m.Hints()) != 0 {
		t.Errorf("no key hints while waiting, got %v", m.Hints())
	}
}

func preferencesThatMove(move Mover) Model {
	return New(config.Default("/app", "/dl")).WithMover(move).SetSize(100, 30)
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

		m, cmd, _, _ := m.Update(keys.Press("m"))
		if m.session.Mode != settings.EditConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.session.Mode, cmd)
		}
		m, _, _, _ = m.Update(cmd())
		if m.session.Mode != settings.EditNone || m.session.Status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.session.Mode, m.session.Status, c.want)
		}
		if view := m.View(140, 30); !strings.Contains(view, c.want[:min(len(c.want), 15)]) {
			t.Errorf("%s: the screen should show the status", c.name)
		}
	}
}

func TestWithoutAMoverThereIsNoMoveKeyAndNoHint(t *testing.T) {
	m := New(config.Default("/app", "/dl")).SetSize(100, 30)

	m, cmd, _, _ := m.Update(keys.Press("m"))
	if cmd != nil || m.session.Mode != settings.EditNone {
		t.Errorf("mode %d, cmd %v: nothing should happen", m.session.Mode, cmd)
	}
	for _, h := range m.Hints() {
		if h.Key == "m" {
			t.Error("the key bar should not offer to move")
		}
	}
}

func TestTheMoveHintIsOfferedBesideUninstall(t *testing.T) {
	m := New(config.Default("/app", "/dl")).
		WithMover(moveAnswering(MoveDeclined, nil)).
		WithUninstaller(uninstallAnswering(RemovalDeclined, nil)).
		SetSize(100, 30)

	var keys []string
	for _, h := range m.Hints() {
		keys = append(keys, h.Key)
	}
	got := strings.Join(keys, " ")
	if !strings.Contains(got, "m u") {
		t.Errorf("hints %q, want move and uninstall side by side", got)
	}
}

func preferencesWith(uninstall Uninstaller) Model {
	return New(config.Default("/app", "/dl")).WithUninstaller(uninstall).SetSize(100, 30)
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

		m, cmd, _, _ := m.Update(keys.Press("u"))
		if m.session.Mode != settings.EditConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.session.Mode, cmd)
		}
		m, _, _, _ = m.Update(cmd())
		if m.session.Mode != settings.EditNone || m.session.Status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.session.Mode, m.session.Status, c.want)
		}
		if got := m.Outline().Prompt; got != c.want {
			t.Errorf("%s: the outline should say it too, got %q", c.name, got)
		}
		if view := m.View(140, 30); !strings.Contains(view, c.want[:20]) {
			t.Errorf("%s: the screen should show the status", c.name)
		}
	}
}

func TestWhileTheQuestionsAreOpenKeysAreIgnored(t *testing.T) {
	m := preferencesWith(uninstallAnswering(RemovalDeclined, nil))
	m, _, _, _ = m.Update(keys.Press("u"))

	m, _, out, _ := m.Update(keys.Press("s"))
	if m.session.Mode != settings.EditConfirming || out != None {
		t.Errorf("mode %d, outcome %d: keys should wait for the answers", m.session.Mode, out)
	}
	if m.Outline().Prompt == "" || len(m.Hints()) != 0 {
		t.Errorf("prompt %q, hints %v", m.Outline().Prompt, m.Hints())
	}
}

func TestAnyKeyClearsTheStatus(t *testing.T) {
	m := preferencesWith(uninstallAnswering(RemovalDeclined, nil))
	m, cmd, _, _ := m.Update(keys.Press("u"))
	m, _, _, _ = m.Update(cmd())
	m, _, _, _ = m.Update(keys.Press("down"))

	if m.session.Status != "" {
		t.Errorf("status %q", m.session.Status)
	}
}

func TestWithoutAnUninstallerThereIsNoKeyAndNoHint(t *testing.T) {
	m := preferencesWith(nil)

	m, cmd, _, _ := m.Update(keys.Press("u"))
	if m.session.Mode != settings.EditNone || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.session.Mode, cmd)
	}
	for _, h := range m.Hints() {
		if h.Key == "u" {
			t.Error("the key bar should not offer uninstall")
		}
	}
}

func TestConfigModelRadioAndSave(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if m.session.Cfg.OfficialVideo != config.OfficialYes {
		t.Fatalf("the official video setting should start as yes, it is %q", m.session.Cfg.OfficialVideo)
	}

	// Move to "Download Official Music Video" (index 6), open it and choose "only".
	for m.session.Cursor < 6 {
		m, _, _, _ = m.Update(keys.Press("down"))
	}
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Mode != settings.EditRadio {
		t.Fatalf("enter should open the choices, mode %d", m.session.Mode)
	}
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.OfficialVideo != config.OfficialOnly {
		t.Errorf("the choice should be only, it is %q", m.session.Cfg.OfficialVideo)
	}

	// The next item saves the songs with no official video.
	for m.session.Cursor < 10 {
		m, _, _, _ = m.Update(keys.Press("down"))
	}
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.SaveNotFound {
		t.Error("enter on the toggle should flip it (it starts as yes)")
	}

	// Save returns the edited config.
	_, _, out, cfg := m.Update(keys.Press("s"))
	if out != Save || cfg.OfficialVideo != config.OfficialOnly || cfg.SaveNotFound {
		t.Errorf("s should save the edited config, out=%d official=%q save=%v", out, cfg.OfficialVideo, cfg.SaveNotFound)
	}
	if _, _, out, _ := m.Update(keys.Press("esc")); out != Cancel {
		t.Errorf("esc should cancel, got %d", out)
	}
}

func TestConfigModelRadio(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	m.session.Cursor = 1 // Video Quality

	m, _, _, _ = m.Update(keys.Press("enter")) // open radio
	if m.session.Mode != settings.EditRadio {
		t.Fatalf("enter on a radio item should open the selector")
	}
	start := m.session.RadioIdx
	m, _, _, _ = m.Update(keys.Press("down"))
	if m.session.RadioIdx != start+1 {
		t.Errorf("down should move the radio cursor")
	}
	m, _, _, _ = m.Update(keys.Press("enter")) // commit
	if m.session.Mode != settings.EditNone {
		t.Error("enter should close the selector")
	}
	if m.session.Cfg.VideoQuality != config.VideoPresets[start+1] {
		t.Errorf("video quality should be %q, got %q", config.VideoPresets[start+1], m.session.Cfg.VideoQuality)
	}
}

func TestConfigModelNumber(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	m.session.Cursor = 5 // Max Concurrent Downloads
	before := m.session.Cfg.MaxConcurrentDownloads

	m, _, _, _ = m.Update(keys.Press("enter")) // edit
	if m.session.Mode != settings.EditNumber {
		t.Fatalf("enter should start number edit")
	}
	m, _, _, _ = m.Update(keys.Press("up")) // increment
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.MaxConcurrentDownloads != before+1 {
		t.Errorf("up then enter should increment to %d, got %d", before+1, m.session.Cfg.MaxConcurrentDownloads)
	}
}

func TestConfigModelTextCommit(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	m.session.Cursor = 4 // Output Template

	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Mode != settings.EditText {
		t.Fatalf("enter should start text edit")
	}
	m.input.SetValue("%(title)s.%(ext)s")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("text edit should commit, got %q", m.session.Cfg.OutputTemplate)
	}
}

func TestConfigModelFolderPicker(t *testing.T) {
	dir := t.TempDir()
	m := New(config.Default(dir, dir))
	m.session.Cursor = 0 // Output Folder -> folder picker

	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Mode != settings.EditFolder {
		t.Fatalf("enter on Output Folder should open the folder picker")
	}
	// Choosing the current folder commits it.
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.session.Mode != settings.EditNone {
		t.Fatalf("enter should choose and close the picker")
	}
	if m.session.Cfg.OutputDir != folderpage.ResolveDir(dir) {
		t.Errorf("output dir should be the chosen folder, got %q", m.session.Cfg.OutputDir)
	}

	// Esc cancels without changing.
	m.session.Cursor = 9 // Log File Location
	m, _, _, _ = m.Update(keys.Press("enter"))
	before := m.session.Cfg.LogDir
	m, _, _, _ = m.Update(keys.Press("esc"))
	if m.session.Mode != settings.EditNone || m.session.Cfg.LogDir != before {
		t.Errorf("esc should cancel the picker without changing the dir")
	}
}

func TestConfigAdvancedTransition(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if _, _, out, _ := m.Update(keys.Press("a")); out != Advanced {
		t.Errorf("'a' should request the advanced screen, got %d", out)
	}
}

func TestConfigColoursTransition(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if _, _, out, _ := m.Update(keys.Press("c")); out != Colours {
		t.Errorf("'c' should request the colours screen, got %d", out)
	}
}
