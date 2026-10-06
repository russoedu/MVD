package localserver

import "net/http"

// ping answers "an MVD lives here", which is how a second start finds the first and how
// an uninstall started from Settings > Apps finds the running app.
func ping(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		App string `json:"app"`
	}{App: "mvd"})
}

// legacyState answers what the page's state route used to, because versions of MVD up to
// 0.0.23 look for a running copy by asking for it: without it, one of those started
// while this one runs would not know and would start a second copy on another port.
func legacyState(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		Version int      `json:"version"`
		Tally   struct{} `json:"tally"`
	}{Version: 1})
}
