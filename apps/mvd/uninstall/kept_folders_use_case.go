package uninstall

import (
	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-server/settings"
)

// keptFolders returns the function that names the folders holding the person's own
// files: the Downloads folder, and the output and log folders in the settings. It reads
// the settings when called, so a change made after the app started is respected.
func keptFolders(configPath, appDir string) func() []string {
	downloads := appdir.DefaultDownloadsDir()
	repository := settings.NewRepository(configPath, appDir, downloads)

	return func() []string {
		keep := []string{downloads}
		if document, err := repository.Load(); err == nil {
			keep = append(keep, document.Settings.OutputDir, document.Settings.LogDir)
		}

		return keep
	}
}
