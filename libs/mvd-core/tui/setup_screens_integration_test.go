package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+o":
		return tea.KeyMsg{Type: tea.KeyCtrlO}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+q":
		return tea.KeyMsg{Type: tea.KeyCtrlQ}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestListModel(t *testing.T) {
	m := newListModel([]string{"https://a", "https://b"})
	if got := m.urls(); len(got) != 2 {
		t.Fatalf("want 2 urls, got %v", got)
	}

	// ctrl+s with entries starts; empty does nothing.
	if _, _, out := m.update(key("ctrl+s")); out != listStart {
		t.Errorf("ctrl+s should start, got %d", out)
	}
	if _, _, out := m.update(key("ctrl+p")); out != listPrefs {
		t.Errorf("ctrl+p should open prefs, got %d", out)
	}

	cleared, _, _ := m.update(key("ctrl+r"))
	if len(cleared.urls()) != 0 {
		t.Errorf("ctrl+r should clear, got %v", cleared.urls())
	}
	if _, _, out := cleared.update(key("ctrl+s")); out != listNone {
		t.Errorf("ctrl+s on empty list should do nothing, got %d", out)
	}
	if _, _, out := cleared.update(key("ctrl+q")); out != listQuit {
		t.Errorf("ctrl+q on empty list should quit, got %d", out)
	}

	// Quit with entries asks to confirm, then y quits.
	confirming, _, out := m.update(key("ctrl+q"))
	if out != listNone || !confirming.confirmQuit {
		t.Fatalf("ctrl+q with entries should ask to confirm")
	}
	if _, _, out := confirming.update(key("y")); out != listQuit {
		t.Errorf("y should confirm quit, got %d", out)
	}
}

func TestConfigModelToggleAndSave(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))

	// Move to "Download Official Music Video" (index 6) and toggle it.
	for m.cursor < 6 {
		m, _, _, _ = m.update(key("down"))
	}
	m, _, _, _ = m.update(key("enter"))
	if !m.cfg.DownloadOfficialMusicVideo {
		t.Error("enter on a toggle should flip it")
	}

	// Save returns the edited config.
	_, _, out, cfg := m.update(key("s"))
	if out != cfgSave || !cfg.DownloadOfficialMusicVideo {
		t.Errorf("s should save the edited config, out=%d official=%v", out, cfg.DownloadOfficialMusicVideo)
	}
	if _, _, out, _ := m.update(key("esc")); out != cfgCancel {
		t.Errorf("esc should cancel, got %d", out)
	}
}

func TestConfigModelRadio(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	m.cursor = 1 // Video Quality

	m, _, _, _ = m.update(key("enter")) // open radio
	if m.mode != editRadio {
		t.Fatalf("enter on a radio item should open the selector")
	}
	start := m.radioIdx
	m, _, _, _ = m.update(key("down"))
	if m.radioIdx != start+1 {
		t.Errorf("down should move the radio cursor")
	}
	m, _, _, _ = m.update(key("enter")) // commit
	if m.mode != editNone {
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
	if m.mode != editNumber {
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
	if m.mode != editText {
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
	if m.mode != editFolder {
		t.Fatalf("enter on Output Folder should open the folder picker")
	}
	// Choosing the current folder commits it.
	m, _, _, _ = m.update(key("enter"))
	if m.mode != editNone {
		t.Fatalf("enter should choose and close the picker")
	}
	if m.cfg.OutputDir != resolveDir(dir) {
		t.Errorf("output dir should be the chosen folder, got %q", m.cfg.OutputDir)
	}

	// Esc cancels without changing.
	m.cursor = 9 // Log File Location
	m, _, _, _ = m.update(key("enter"))
	before := m.cfg.LogDir
	m, _, _, _ = m.update(key("esc"))
	if m.mode != editNone || m.cfg.LogDir != before {
		t.Errorf("esc should cancel the picker without changing the dir")
	}
}

func TestConfigAdvancedTransition(t *testing.T) {
	m := newConfigModel(config.Default("/app", "/dl"))
	if _, _, out, _ := m.update(key("a")); out != cfgAdvanced {
		t.Errorf("'a' should request the advanced screen, got %d", out)
	}
}

func TestAdvancedModel(t *testing.T) {
	m := newAdvancedModel(config.Default("/app", "/dl"))

	// Toggle auto retry (item 2).
	m.cursor = 2
	before := m.cfg.AutoRetry
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.AutoRetry == before {
		t.Error("enter should toggle auto retry")
	}

	// Fragment count (item 1): edit, decrement, commit.
	m.cursor = 1
	start := m.cfg.ConcurrentFragments
	m, _, _, _ = m.update(key("enter"))
	if m.mode != editNumber {
		t.Fatalf("enter should edit the fragment count")
	}
	m, _, _, _ = m.update(key("down"))
	m, _, _, _ = m.update(key("enter"))
	if m.cfg.ConcurrentFragments != start-1 {
		t.Errorf("fragments should decrement to %d, got %d", start-1, m.cfg.ConcurrentFragments)
	}

	// Save returns the edited config.
	_, _, out, cfg := m.update(key("s"))
	if out != advSave || cfg.AutoRetry == before {
		t.Errorf("s should save the edited config")
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
