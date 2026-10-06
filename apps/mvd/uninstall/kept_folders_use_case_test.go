package uninstall

import (
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
)

func TestTheOutputAndLogFoldersFromTheSettingsAreKeptAlongsideDownloads(t *testing.T) {
	appDir := t.TempDir()
	configPath := filepath.Join(appDir, "config.conf")
	output, logs := t.TempDir(), t.TempDir()
	cfg := config.Default(appDir, appdir.DefaultDownloadsDir())
	cfg.OutputDir = output
	cfg.CreateLogFile = true
	cfg.LogDir = logs
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	keep := keptFolders(configPath, appDir)()

	for _, want := range []string{appdir.DefaultDownloadsDir(), output, logs} {
		found := false
		for _, got := range keep {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s is not among the kept folders %v", want, keep)
		}
	}
}

func TestWithoutSettingsTheDownloadsFolderIsStillKept(t *testing.T) {
	appDir := t.TempDir()

	keep := keptFolders(filepath.Join(appDir, "missing.conf"), appDir)()

	if len(keep) == 0 || keep[0] != appdir.DefaultDownloadsDir() {
		t.Errorf("keep = %v", keep)
	}
}
