package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNoDialog is what a FolderPicker returns on a machine that cannot show a chooser
// (no desktop, or none of the tools it relies on).
var ErrNoDialog = errors.New("this machine has no folder chooser")

// folderPickRequest starts the chooser at a folder the page already knows.
type folderPickRequest struct {
	Start string `json:"start"`
}

// folderPickResponse is the folder chosen, or Cancelled when the person closed the
// chooser without choosing.
type folderPickResponse struct {
	Path      string `json:"path"`
	Cancelled bool   `json:"cancelled"`
}

// pickFolder opens a chooser on the machine MVD runs on and waits for the answer.
//
// The server is on the same machine as the browser, so this is the person's own
// screen. Only one chooser is open at a time: a second request while one is showing
// is refused, so a stuck page cannot stack windows.
func (h *handler) pickFolder(w http.ResponseWriter, r *http.Request) {
	var body folderPickRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSettingsBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "expected {\"start\": \"...\"}"})
		return
	}

	select {
	case h.pick <- struct{}{}:
		defer func() { <-h.pick }()
	default:
		writeJSON(w, http.StatusConflict, errorBody{Error: "a folder chooser is already open"})
		return
	}

	path, chosen, err := h.folders.Pick(r.Context(), body.Start)
	switch {
	case errors.Is(err, ErrNoDialog):
		writeJSON(w, http.StatusNotImplemented, errorBody{Error: "no folder chooser is available here; type the path instead"})
	case err != nil:
		writeJSON(w, http.StatusBadGateway, errorBody{Error: "the folder chooser failed: " + err.Error()})
	case !chosen:
		writeJSON(w, http.StatusOK, folderPickResponse{Cancelled: true})
	default:
		writeJSON(w, http.StatusOK, folderPickResponse{Path: path})
	}
}
