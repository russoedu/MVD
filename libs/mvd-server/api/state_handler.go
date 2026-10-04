package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// state serves the run as it is now.
func (h *handler) state(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.sessions.Snapshot())
}

// events streams the run as server-sent events: the whole snapshot on connecting,
// and again each time it changes, no more often than the throttle allows.
//
// Snapshots, not individual changes, so a client that missed some (a laptop that
// slept, a reconnect) needs no catching up and has no reducer to keep in step with
// the engine's: it just draws what arrives.
func (h *handler) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "streaming is not supported here"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	// Tells a reverse proxy not to hold the stream back until it fills a buffer.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// -1 is older than any real version, so the first wait returns at once.
	version := int64(-1)
	for {
		waiting, cancel := context.WithTimeout(r.Context(), h.keepalive)
		snap, changed := h.sessions.WaitSince(waiting, version)
		cancel()
		if r.Context().Err() != nil {
			return
		}
		if !changed {
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
			continue
		}

		body, err := json.Marshal(snap)
		if err != nil {
			return
		}
		if _, err := fmt.Fprintf(w, "id: %d\nevent: snapshot\ndata: %s\n\n", snap.Version, body); err != nil {
			return
		}
		flusher.Flush()
		version = snap.Version

		select {
		case <-time.After(h.throttle):
		case <-r.Context().Done():
			return
		}
	}
}
