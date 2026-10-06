package localserver

import (
	"net/http"
	"path/filepath"
)

// ConfigPath is the settings file, the same one the terminal app uses.
func ConfigPath(appDir string) string {
	return filepath.Join(appDir, "config.conf")
}

// TerminalPath is where the terminal interface's WebSocket is served, when there is one.
const TerminalPath = "/term"

// NewHandler is the whole site: the few routes under /api/ (the identity of the app, and
// uninstalling it when uninstaller is not nil), the terminal interface's WebSocket under
// /term when terminal is not nil, and the built page for everything else.
func NewHandler(uninstaller Uninstaller, terminal http.Handler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/ping", ping)
	api.HandleFunc("GET /api/state", legacyState)
	if uninstaller != nil {
		api.Handle("POST /api/uninstall", newUninstallHandler(uninstaller))
	}

	mux := http.NewServeMux()
	// Every route under /api/ goes through the request guard first, so nothing there can
	// be reached by a page on another site.
	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reason := refusal(r); reason != "" {
			writeJSON(w, http.StatusForbidden, errorBody{Error: reason})
			return
		}
		api.ServeHTTP(w, r)
	}))
	if terminal != nil {
		mux.Handle(TerminalPath, terminal)
	}
	mux.Handle("/", webHandler())

	return mux
}
