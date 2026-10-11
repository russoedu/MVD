package tui

import (
	"testing"
	"youtube-downloader/libs/mvd-core/settings"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"
	"youtube-downloader/libs/mvd-core/testing/keys"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func key(s string) tea.KeyPressMsg { return keys.Press(s) }

func TestConfigModelRadioAndSave(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	if m.cfg.OfficialVideo != config.OfficialYes {
		t.Fatalf("the official video setting should start as yes, it is %q", m.cfg.OfficialVideo)
	}

	// Move to "Download Official Music Video" (index 6), open it and choose "only".
	for m.cursor < 6 {
		m, _, _, _ = m.update(key("down"))
	}
	m, _, _, _ = m.update(key("enter"))
	if m.mode != settings.EditRadio {
		t.Fatalf("enter should open the choices, mode %d", m.mode)
	}
	m, _, _, _ = m.update(key("down"))
	m, _, _, _ = m.update(key("down"))
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.OfficialVideo != config.OfficialOnly {
		t.Errorf("the choice should be only, it is %q", m.cfg.OfficialVideo)
	}

	// The next item saves the songs with no official video.
	for m.cursor < 10 {
		m, _, _, _ = m.update(key("down"))
	}
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.SaveNotFound {
		t.Error("enter on the toggle should flip it (it starts as yes)")
	}

	// Save returns the edited config.
	_, _, out, cfg := m.update(key("s"))
	if out != cfgSave || cfg.OfficialVideo != config.OfficialOnly || cfg.SaveNotFound {
		t.Errorf("s should save the edited config, out=%d official=%q save=%v", out, cfg.OfficialVideo, cfg.SaveNotFound)
	}
	if _, _, out, _ := m.update(key("esc")); out != cfgCancel {
		t.Errorf("esc should cancel, got %d", out)
	}
}

func TestConfigModelRadio(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	m.cursor = 1 // Video Quality

	m, _, _, _ = m.update(key("enter")) // open radio
	if m.mode != settings.EditRadio {
		t.Fatalf("enter on a radio item should open the selector")
	}
	start := m.radioIdx
	m, _, _, _ = m.update(key("down"))
	if m.radioIdx != start+1 {
		t.Errorf("down should move the radio cursor")
	}
	m, _, _, _ = m.update(key("enter")) // commit
	if m.mode != settings.EditNone {
		t.Error("enter should close the selector")
	}
	if m.cfg.VideoQuality != config.VideoPresets[start+1] {
		t.Errorf("video quality should be %q, got %q", config.VideoPresets[start+1], m.cfg.VideoQuality)
	}
}

func TestConfigModelNumber(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	m.cursor = 5 // Max Concurrent Downloads
	before := m.cfg.MaxConcurrentDownloads

	m, _, _, _ = m.update(key("enter")) // edit
	if m.mode != settings.EditNumber {
		t.Fatalf("enter should start number edit")
	}
	m, _, _, _ = m.update(key("up")) // increment
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.MaxConcurrentDownloads != before+1 {
		t.Errorf("up then enter should increment to %d, got %d", before+1, m.cfg.MaxConcurrentDownloads)
	}
}

func TestConfigModelTextCommit(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	m.cursor = 4 // Output Template

	m, _, _, _ = m.update(key("enter"))
	if m.mode != settings.EditText {
		t.Fatalf("enter should start text edit")
	}
	m.input.SetValue("%(title)s.%(ext)s")
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.OutputTemplate != "%(title)s.%(ext)s" {
		t.Errorf("text edit should commit, got %q", m.cfg.OutputTemplate)
	}
}

func TestConfigModelFolderPicker(t *testing.T) {
	dir := t.TempDir()
	m := newConfigModel(config.Default(dir, dir))
	m.cursor = 0 // Output Folder -> folder picker

	m, _, _, _ = m.update(key("enter"))
	if m.mode != settings.EditFolder {
		t.Fatalf("enter on Output Folder should open the folder picker")
	}
	// Choosing the current folder commits it.
	m, _, _, _ = m.update(key("enter"))
	if m.mode != settings.EditNone {
		t.Fatalf("enter should choose and close the picker")
	}
	if m.cfg.OutputDir != folderpage.ResolveDir(dir) {
		t.Errorf("output dir should be the chosen folder, got %q", m.cfg.OutputDir)
	}

	// Esc cancels without changing.
	m.cursor = 9 // Log File Location
	m, _, _, _ = m.update(key("enter"))
	before := m.cfg.LogDir
	m, _, _, _ = m.update(key("esc"))
	if m.mode != settings.EditNone || m.cfg.LogDir != before {
		t.Errorf("esc should cancel the picker without changing the dir")
	}
}

func TestConfigAdvancedTransition(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	if _, _, out, _ := m.update(key("a")); out != cfgAdvanced {
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

// keyText is the key press that types text.
func keyText(text string) tea.KeyPressMsg { return keys.Text(text) }

func TestConfigColoursTransition(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	if _, _, out, _ := m.update(key("c")); out != cfgColours {
		t.Errorf("'c' should request the colours screen, got %d", out)
	}
}
