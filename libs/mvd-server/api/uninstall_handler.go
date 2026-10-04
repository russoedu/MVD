package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// uninstallRequest says whether the stored preferences go too. Left out, the person is
// asked on their own screen.
type uninstallRequest struct {
	DeletePreferences *bool `json:"deletePreferences"`
}

// uninstallResponse tells the page the removal has been accepted and is under way.
type uninstallResponse struct {
	Status string `json:"status"`
}

// uninstall asks the machine's owner to confirm removing the app, and removes it once
// they have.
//
// The route can destroy something and is reachable from any page that can talk to
// localhost, so the request guard is not the only protection: the confirmation is
// always put to the person on their own screen, whatever the page said, and a "no"
// there ends the request without removing anything. The removal itself runs after the
// answer has been sent, because it ends with the app quitting.
//
// Only one confirmation is open at a time.
func (h *handler) uninstall(w http.ResponseWriter, r *http.Request) {
	var body uninstallRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSettingsBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "expected {\"deletePreferences\": true|false}"})
		return
	}

	select {
	case h.confirming <- struct{}{}:
		defer func() { <-h.confirming }()
	default:
		writeJSON(w, http.StatusConflict, errorBody{Error: "a removal is already being confirmed"})
		return
	}

	remove, err := h.uninstaller.Confirm(body.DeletePreferences)
	switch {
	case errors.Is(err, ErrUninstallDeclined):
		writeJSON(w, http.StatusConflict, errorBody{Error: "the removal was cancelled; nothing was removed"})
	case errors.Is(err, ErrNoDialog):
		writeJSON(w, http.StatusNotImplemented, errorBody{Error: "this machine cannot ask for confirmation, so nothing was removed; start MVD with -uninstall from a terminal"})
	case err != nil:
		writeJSON(w, http.StatusBadGateway, errorBody{Error: "the removal could not be confirmed: " + err.Error()})
	default:
		writeJSON(w, http.StatusAccepted, uninstallResponse{Status: "removing"})
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		go remove()
	}
}
