package config

import (
	"path/filepath"
	"testing"
)

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
