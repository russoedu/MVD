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

	if m.mode != settings.EditPicking || cmd == nil {
		t.Fatalf("mode %d, cmd %v: the chooser should be open", m.mode, cmd)
	}
	answer := cmd()
	if startedAt != "/dl" {
		t.Errorf("the chooser should start at the current folder, started at %q", startedAt)
	}

	m, _, _, _ = m.Update(answer)
	if m.mode != settings.EditNone || m.cfg.OutputDir != "/music" {
		t.Errorf("mode %d, output dir %q", m.mode, m.cfg.OutputDir)
	}
}

func TestCancellingTheChooserChangesNothing(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })
	m, _, _, _ = m.Update(cmd())

	if m.mode != settings.EditNone || m.cfg.OutputDir != "/dl" {
		t.Errorf("mode %d, output dir %q", m.mode, m.cfg.OutputDir)
	}
}

func TestWithoutAChooserOnTheMachineTheBuiltInBrowserOpens(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, errors.New("no desktop") })
	m, _, _, _ = m.Update(cmd())

	if m.mode != settings.EditFolder {
		t.Errorf("mode %d, want the built-in folder browser", m.mode)
	}
}

func TestWithoutAHostChooserTheBuiltInBrowserOpensAtOnce(t *testing.T) {
	m, cmd := chooseOnFirstSetting(t, nil)

	if m.mode != settings.EditFolder || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.mode, cmd)
	}
}

func TestTheLogFolderUsesTheChooserToo(t *testing.T) {
	m := New(config.Default("/app", "/dl")).WithFolderPicker(func(start string) (string, bool, error) {
		return "/logs", true, nil
	}).SetSize(100, 30)
	m.cursor = 9
	m, cmd, _, _ := m.Update(keys.Press("enter"))
	m, _, _, _ = m.Update(cmd())

	if m.cfg.LogDir != "/logs" {
		t.Errorf("log dir %q", m.cfg.LogDir)
	}
}

func TestWhileTheChooserIsOpenKeysAreIgnoredAndTheScreenSaysWhy(t *testing.T) {
	m, _ := chooseOnFirstSetting(t, func(string) (string, bool, error) { return "", false, nil })

	m, _, _, _ = m.Update(keys.Press("esc"))
	m, _, out, _ := m.Update(keys.Press("s"))
	if m.mode != settings.EditPicking || out != None {
		t.Errorf("mode %d, outcome %d: keys should wait for the chooser", m.mode, out)
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
		if m.mode != settings.EditConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.mode, cmd)
		}
		m, _, _, _ = m.Update(cmd())
		if m.mode != settings.EditNone || m.status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.mode, m.status, c.want)
		}
		if view := m.View(140, 30); !strings.Contains(view, c.want[:min(len(c.want), 15)]) {
			t.Errorf("%s: the screen should show the status", c.name)
		}
	}
}

func TestWithoutAMoverThereIsNoMoveKeyAndNoHint(t *testing.T) {
	m := New(config.Default("/app", "/dl")).SetSize(100, 30)

	m, cmd, _, _ := m.Update(keys.Press("m"))
	if cmd != nil || m.mode != settings.EditNone {
		t.Errorf("mode %d, cmd %v: nothing should happen", m.mode, cmd)
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
		if m.mode != settings.EditConfirming || cmd == nil {
			t.Fatalf("%s: mode %d, cmd %v: the questions should be open", c.name, m.mode, cmd)
		}
		m, _, _, _ = m.Update(cmd())
		if m.mode != settings.EditNone || m.status != c.want {
			t.Errorf("%s: mode %d, status %q, want %q", c.name, m.mode, m.status, c.want)
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
	if m.mode != settings.EditConfirming || out != None {
		t.Errorf("mode %d, outcome %d: keys should wait for the answers", m.mode, out)
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

	if m.status != "" {
		t.Errorf("status %q", m.status)
	}
}

func TestWithoutAnUninstallerThereIsNoKeyAndNoHint(t *testing.T) {
	m := preferencesWith(nil)

	m, cmd, _, _ := m.Update(keys.Press("u"))
	if m.mode != settings.EditNone || cmd != nil {
		t.Errorf("mode %d, cmd %v", m.mode, cmd)
	}
	for _, h := range m.Hints() {
		if h.Key == "u" {
			t.Error("the key bar should not offer uninstall")
		}
	}
}

func TestConfigModelRadioAndSave(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if m.cfg.OfficialVideo != config.OfficialYes {
		t.Fatalf("the official video setting should start as yes, it is %q", m.cfg.OfficialVideo)
	}

	// Move to "Download Official Music Video" (index 6), open it and choose "only".
	for m.cursor < 6 {
		m, _, _, _ = m.Update(keys.Press("down"))
	}
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.mode != settings.EditRadio {
		t.Fatalf("enter should open the choices, mode %d", m.mode)
	}
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("down"))
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.OfficialVideo != config.OfficialOnly {
		t.Errorf("the choice should be only, it is %q", m.cfg.OfficialVideo)
	}

	// The next item saves the songs with no official video.
	for m.cursor < 10 {
		m, _, _, _ = m.Update(keys.Press("down"))
	}
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.SaveNotFound {
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
	m.cursor = 1 // Video Quality

	m, _, _, _ = m.Update(keys.Press("enter")) // open radio
	if m.mode != settings.EditRadio {
		t.Fatalf("enter on a radio item should open the selector")
	}
	start := m.radioIdx
	m, _, _, _ = m.Update(keys.Press("down"))
	if m.radioIdx != start+1 {
		t.Errorf("down should move the radio cursor")
	}
	m, _, _, _ = m.Update(keys.Press("enter")) // commit
	if m.mode != settings.EditNone {
		t.Error("enter should close the selector")
	}
	if m.cfg.VideoQuality != config.VideoPresets[start+1] {
		t.Errorf("video quality should be %q, got %q", config.VideoPresets[start+1], m.cfg.VideoQuality)
	}
}

func TestConfigModelNumber(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	m.cursor = 5 // Max Concurrent Downloads
	before := m.cfg.MaxConcurrentDownloads

	m, _, _, _ = m.Update(keys.Press("enter")) // edit
	if m.mode != settings.EditNumber {
		t.Fatalf("enter should start number edit")
	}
	m, _, _, _ = m.Update(keys.Press("up")) // increment
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.MaxConcurrentDownloads != before+1 {
		t.Errorf("up then enter should increment to %d, got %d", before+1, m.cfg.MaxConcurrentDownloads)
	}
}

func TestConfigModelTextCommit(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	m.cursor = 4 // Output Template

	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.mode != settings.EditText {
		t.Fatalf("enter should start text edit")
	}
	m.input.SetValue("%(title)s.%(ext)s")
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("text edit should commit, got %q", m.cfg.OutputTemplate)
	}
}

func TestConfigModelFolderPicker(t *testing.T) {
	dir := t.TempDir()
	m := New(config.Default(dir, dir))
	m.cursor = 0 // Output Folder -> folder picker

	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.mode != settings.EditFolder {
		t.Fatalf("enter on Output Folder should open the folder picker")
	}
	// Choosing the current folder commits it.
	m, _, _, _ = m.Update(keys.Press("enter"))
	if m.mode != settings.EditNone {
		t.Fatalf("enter should choose and close the picker")
	}
	if m.cfg.OutputDir != folderpage.ResolveDir(dir) {
		t.Errorf("output dir should be the chosen folder, got %q", m.cfg.OutputDir)
	}

	// Esc cancels without changing.
	m.cursor = 9 // Log File Location
	m, _, _, _ = m.Update(keys.Press("enter"))
	before := m.cfg.LogDir
	m, _, _, _ = m.Update(keys.Press("esc"))
	if m.mode != settings.EditNone || m.cfg.LogDir != before {
		t.Errorf("esc should cancel the picker without changing the dir")
	}
}

func TestConfigAdvancedTransition(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if _, _, out, _ := m.Update(keys.Press("a")); out != Advanced {
		t.Errorf("'a' should request the advanced screen, got %d", out)
	}
}

func TestCookieChoice(t *testing.T) {
	cfg := config.Default("/app", "/dl")
	applyCookieChoice(&cfg, "firefox")
	if cfg.CookiesFromBrowser != "firefox" || cfg.AutoCookies {
		t.Errorf("pin wrong: %+v", cfg)
	}
	if currentCookieChoice(cfg) != "firefox" {
		t.Error("currentCookieChoice should report the pin")
	}
	applyCookieChoice(&cfg, "off")
	if cfg.AutoCookies || cfg.CookiesFromBrowser != "" || currentCookieChoice(cfg) != "off" {
		t.Errorf("off wrong: %+v", cfg)
	}
	applyCookieChoice(&cfg, "all")
	if !cfg.AutoCookies || currentCookieChoice(cfg) != "all" {
		t.Errorf("all wrong: %+v", cfg)
	}
}

func TestConfigColoursTransition(t *testing.T) {
	m := New(config.Default("/app", "/dl"))
	if _, _, out, _ := m.Update(keys.Press("c")); out != Colours {
		t.Errorf("'c' should request the colours screen, got %d", out)
	}
}
