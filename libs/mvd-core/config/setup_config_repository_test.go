package config

import (
	"os"
	"path/filepath"
	"testing"
)

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
	cfg.OfficialVideo = OfficialOnly
	cfg.SaveNotFound = false
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
		reloaded.MaxConcurrentDownloads != 7 || reloaded.OfficialVideo != OfficialOnly || reloaded.SaveNotFound ||
		reloaded.CookiesFromBrowser != "firefox" || reloaded.AutoCookies ||
		reloaded.CreateLogFile || len(reloaded.ExtraArgs) != 2 {
		t.Errorf("reloaded config wrong: %+v", reloaded)
	}
	if reloaded.Format() != "bestvideo[height<=1080]+bestaudio[abr<=128]/best[height<=1080]" {
		t.Errorf("compiled format wrong: %q", reloaded.Format())
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
