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

func TestLoadSetup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "setup.conf")
	content := "download_official_music_video = true\nquality=best\nlog_file=off\nmax_concurrent_downloads=5\nextra_args=-4 --js-runtimes node\n# comment\nbroken line\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadSetup(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DownloadOfficialMusicVideo || cfg.Quality != "best" || cfg.LogFile != "" || cfg.MaxConcurrentDownloads != 5 || len(cfg.ExtraArgs) != 3 {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if Default().DownloadOfficialMusicVideo || Default().LogFile != "mvd.log" {
		t.Error("unexpected defaults")
	}

	missing, err := LoadSetup(filepath.Join(dir, "nope.conf"))
	if err != nil || missing.Quality != Default().Quality {
		t.Errorf("missing file should yield defaults, got %+v, %v", missing, err)
	}
}

func TestLoadDownloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "downloads.conf")
	if err := os.WriteFile(path, []byte("# list\nhttps://a\n\n  https://b  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	urls, err := LoadDownloads(path)
	if err != nil || len(urls) != 2 || urls[0] != "https://a" || urls[1] != "https://b" {
		t.Errorf("unexpected urls %v, %v", urls, err)
	}
}
