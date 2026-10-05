package localserver

import (
	"net/http"
	"path/filepath"

	"youtube-downloader/libs/mvd-server/api"
)

// ConfigPath is the settings file, the same one the terminal app uses.
func ConfigPath(appDir string) string {
	return filepath.Join(appDir, "config.conf")
}

// TerminalPath is where the terminal interface's WebSocket is served, when there is one.
const TerminalPath = "/term"

// NewHandler is the whole site: the API under /api/, the terminal interface's
// WebSocket under /term when terminal is not nil, and the built frontend for
// everything else.
func NewHandler(sessions api.Sessions, settings api.SettingsStore, folders api.FolderPicker, uninstaller api.Uninstaller, terminal http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", api.New(sessions, api.Options{Settings: settings, Folders: folders, Uninstall: uninstaller}))
	if terminal != nil {
		mux.Handle(TerminalPath, terminal)
	}
	mux.Handle("/", webHandler())

	return mux
}
