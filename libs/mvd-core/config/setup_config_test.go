package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBool(t *testing.T) {
	for _, v := range []string{"true", "True", "yes", "1", "on"} {
		if !parseBool(v) {
			t.Errorf("%q should be true", v)
		}
	}
	for _, v := range []string{"false", "0", "no", "", "maybe"} {
		if parseBool(v) {
			t.Errorf("%q should be false", v)
		}
	}
}

func TestDefaults(t *testing.T) {
	d := Default("/app", "/downloads")
	if d.OutputDir != "/downloads" {
		t.Errorf("output dir should default to downloads, got %q", d.OutputDir)
	}
	if d.VideoQuality != "best" || d.AudioQuality != "best" || d.RawFormat != "" {
		t.Errorf("quality defaults wrong: %+v", d)
	}
	if d.MaxConcurrentDownloads != 4 || d.ConcurrentFragments != 4 || !d.AutoRetry || !d.AutoCookies {
		t.Errorf("unexpected defaults: %+v", d)
	}
	if d.CookiesFile != filepath.Join("/app", "cookies.txt") || d.LogDir != "/app" || !d.CreateLogFile {
		t.Errorf("path defaults wrong: %+v", d)
	}
	if d.Format() != "bestvideo+bestaudio/best" {
		t.Errorf("default format wrong: %q", d.Format())
	}
	if got := d.LogFile(); got != filepath.Join("/app", "mvd-downloads.log") {
		t.Errorf("log file wrong: %q", got)
	}
}

func TestLoadOrCreateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.conf")

	cfg, created, err := LoadOrCreate(path, dir, filepath.Join(dir, "dl"))
	if err != nil || !created {
		t.Fatalf("first load should create: created=%v err=%v", created, err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("config file not written: %v", statErr)
	}

	// Second load does not report created and reads the saved values.
	again, created2, err := LoadOrCreate(path, dir, filepath.Join(dir, "dl"))
	if err != nil || created2 {
		t.Fatalf("second load should not create: created=%v err=%v", created2, err)
	}
	if again.OutputDir != cfg.OutputDir || again.Format() != cfg.Format() {
		t.Errorf("round-trip mismatch: %+v vs %+v", again, cfg)
	}

	// Modify, save, reload.
	cfg.VideoQuality = "1080p"
	cfg.AudioQuality = "medium"
	cfg.MaxConcurrentDownloads = 7
	cfg.DownloadOfficialMusicVideo = true
	cfg.CookiesFromBrowser = "firefox"
	cfg.AutoCookies = false
	cfg.CreateLogFile = false
	cfg.ExtraArgs = []string{"-4", "--no-warnings"}
	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}
	reloaded, _, err := LoadOrCreate(path, dir, filepath.Join(dir, "dl"))
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.VideoQuality != "1080p" || reloaded.AudioQuality != "medium" ||
		reloaded.MaxConcurrentDownloads != 7 || !reloaded.DownloadOfficialMusicVideo ||
		reloaded.CookiesFromBrowser != "firefox" || reloaded.AutoCookies ||
		reloaded.CreateLogFile || len(reloaded.ExtraArgs) != 2 {
		t.Errorf("reloaded config wrong: %+v", reloaded)
	}
	if reloaded.Format() != "bestvideo[height<=1080]+bestaudio[abr<=128]/best[height<=1080]" {
		t.Errorf("compiled format wrong: %q", reloaded.Format())
	}
}

func TestCookieSettingRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.conf")
	base := Default(dir, dir)
	for in, want := range map[string]struct {
		auto bool
		pin  string
	}{
		"all":     {true, ""},
		"off":     {false, ""},
		"firefox": {false, "firefox"},
	} {
		c := base
		applyKey(&c, "cookies_from_browser", in, path)
		if c.AutoCookies != want.auto || c.CookiesFromBrowser != want.pin {
			t.Errorf("%q -> auto=%v pin=%q", in, c.AutoCookies, c.CookiesFromBrowser)
		}
		if got := cookieSetting(c); got != in {
			t.Errorf("round-trip %q -> %q", in, got)
		}
	}
}

func TestLegacyKeys(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "setup.conf")
	if err := os.WriteFile(legacy, []byte("quality=bestvideo+bestaudio/best\nlog_file=off\ncookies_from_browser=edge\nmax_concurrent_downloads=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseInto(legacy, Default(dir, dir))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RawFormat != "bestvideo+bestaudio/best" {
		t.Errorf("legacy quality should become raw format, got %q", cfg.RawFormat)
	}
	if cfg.CreateLogFile {
		t.Error("log_file=off should turn logging off")
	}
	if cfg.CookiesFromBrowser != "edge" || cfg.AutoCookies {
		t.Errorf("legacy cookie pin wrong: %+v", cfg)
	}
	if cfg.MaxConcurrentDownloads != 2 {
		t.Errorf("legacy concurrency wrong: %d", cfg.MaxConcurrentDownloads)
	}
}
