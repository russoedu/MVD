package api

import (
	"net/http"
	"strconv"
)

// retryEntry re-queues one failed entry.
func (h *handler) retryEntry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "the entry id must be a number"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": h.sessions.Retry(id)})
}

// retryPlaylist re-queues the failed entries of one playlist.
func (h *handler) retryPlaylist(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "the playlist index must be a number"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"retried": h.sessions.RetryPlaylist(index)})
}
