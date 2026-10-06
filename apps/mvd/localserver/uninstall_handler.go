package localserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"youtube-downloader/apps/mvd/uninstall"
)

// maxUninstallBody is the most the uninstall request may send.
const maxUninstallBody = 1 << 10

// Uninstaller removes the app from the machine, after the person has agreed.
type Uninstaller interface {
	// Confirm puts what will be removed to the person on the machine's own screen and
	// waits. deletePreferences says whether the stored preferences go too, or nil to ask
	// the person as well. It returns uninstall.ErrDeclined when they say no and
	// uninstall.ErrNoDialog when there is no way to ask. Otherwise it returns the removal,
	// which the caller runs once it has answered: it ends with the app quitting.
	Confirm(deletePreferences *bool) (remove func(), err error)
}

// uninstallRequest says whether the stored preferences go too. Left out, the person is
// asked on their own screen.
type uninstallRequest struct {
	DeletePreferences *bool `json:"deletePreferences"`
}

type uninstallResponse struct {
	Status string `json:"status"`
}

// uninstallHandler asks the machine's owner to confirm removing the app, and removes it
// once they have.
//
// The route can destroy something and is reachable from any page that can talk to
// localhost, so the request guard is not the only protection: the confirmation is
// always put to the person on their own screen, whatever the request said, and a "no"
// there ends the request without removing anything. The removal itself runs after the
// answer has been sent, because it ends with the app quitting.
//
// Only one confirmation is open at a time.
type uninstallHandler struct {
	uninstaller Uninstaller
	confirming  chan struct{}
}

func newUninstallHandler(uninstaller Uninstaller) http.Handler {
	return &uninstallHandler{uninstaller: uninstaller, confirming: make(chan struct{}, 1)}
}

func (h *uninstallHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body uninstallRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxUninstallBody))
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
	case errors.Is(err, uninstall.ErrDeclined):
		writeJSON(w, http.StatusConflict, errorBody{Error: "the removal was cancelled; nothing was removed"})
	case errors.Is(err, uninstall.ErrNoDialog):
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
