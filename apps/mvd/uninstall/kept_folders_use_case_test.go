package uninstall

import (
	"path/filepath"
	"testing"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-server/settings"
)

func TestTheOutputAndLogFoldersFromTheSettingsAreKeptAlongsideDownloads(t *testing.T) {
	appDir := t.TempDir()
	configPath := filepath.Join(appDir, "config.conf")
	output, logs := t.TempDir(), t.TempDir()
	repository := settings.NewRepository(configPath, appDir, appdir.DefaultDownloadsDir())
	document, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}
	document.Settings.OutputDir = output
	document.Settings.CreateLogFile = true
	document.Settings.LogDir = logs
	if err := repository.Save(document.Settings); err != nil {
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
