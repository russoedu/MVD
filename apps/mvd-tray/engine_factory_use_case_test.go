package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheFactoryReadsTheSettingsWhenItIsCalledAndCreatesTheDownloadFolder(t *testing.T) {
	appDir := t.TempDir()
	out := filepath.Join(t.TempDir(), "music", "videos")
	// Cookies off, so building the engine never goes looking for a browser or runs yt-dlp.
	config := "output_dir=" + out + "\ncookies_from_browser=off\ncookies_file=off\ncreate_log_file=false\n"
	if err := os.WriteFile(filepath.Join(appDir, "config.conf"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eng, err := newEngineFactory("yt-dlp", appDir, nil)(ctx, []string{"https://a.example/p"})

	if err != nil {
		t.Fatal(err)
	}
	if eng == nil {
		t.Fatal("no engine")
	}
	if info, statErr := os.Stat(out); statErr != nil || !info.IsDir() {
		t.Errorf("the download folder was not created: %v", statErr)
	}
}

func TestAnUnreadableSettingsFileIsReportedNotSwallowed(t *testing.T) {
	appDir := t.TempDir()
	// A directory where the file should be cannot be read as one.
	if err := os.Mkdir(filepath.Join(appDir, "config.conf"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := newEngineFactory("yt-dlp", appDir, nil)(context.Background(), nil)

	if err == nil || !strings.Contains(err.Error(), "settings") {
		t.Errorf("err = %v", err)
	}
}
