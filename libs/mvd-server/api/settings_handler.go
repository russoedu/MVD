package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"youtube-downloader/libs/mvd-server/settings"
)

// maxSettingsBody is far beyond any real settings and keeps a runaway client from
// filling memory.
const maxSettingsBody = 64 << 10

// getSettings serves the settings and the choices for them.
func (h *handler) getSettings(w http.ResponseWriter, _ *http.Request) {
	doc, err := h.settings.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "cannot read the settings: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// putSettings replaces the settings. A value that is not acceptable gets a 400 that
// names the field, and nothing is written.
func (h *handler) putSettings(w http.ResponseWriter, r *http.Request) {
	var body settings.Settings
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSettingsBody))
	// Unknown fields are an error so a misspelt key is not silently left unchanged.
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "expected the settings as a JSON object"})
		return
	}

	err := h.settings.Save(body)

	var invalid settings.Invalid
	switch {
	case errors.As(err, &invalid):
		writeJSON(w, http.StatusBadRequest, errorBody{Error: invalid.Error(), Fields: invalid})
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "cannot save the settings: " + err.Error()})
	default:
		h.getSettings(w, r)
	}
}
