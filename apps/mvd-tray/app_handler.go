package main

import (
	"net/http"
	"path/filepath"

	"youtube-downloader/libs/mvd-server/api"
)

// configPath is the settings file, the same one the terminal app uses.
func configPath(appDir string) string {
	return filepath.Join(appDir, "config.conf")
}

// newAppHandler is the whole site: the API under /api/ and the built frontend for
// everything else.
func newAppHandler(sessions api.Sessions, settings api.SettingsStore, folders api.FolderPicker) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", api.New(sessions, api.Options{Settings: settings, Folders: folders}))
	mux.Handle("/", webHandler())

	return mux
}
