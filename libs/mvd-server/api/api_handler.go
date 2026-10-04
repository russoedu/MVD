package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// Options tune the handler. The zero value is what the app uses.
type Options struct {
	// Throttle is the shortest gap between two pushes on the event stream. A busy
	// download changes the run many times a second and a browser gains nothing from
	// every one, so changes arriving inside the gap are sent as one. Default 250ms.
	Throttle time.Duration
	// Keepalive is how long the stream may stay silent before it sends a comment, so
	// a proxy or the browser does not decide it has died. Default 15s.
	Keepalive time.Duration
}

type handler struct {
	sessions  Sessions
	throttle  time.Duration
	keepalive time.Duration
}

// New returns the API, to be mounted at /api/. Every route goes through the request
// guard first, so nothing here can be reached by a page on another site.
func New(sessions Sessions, options Options) http.Handler {
	h := &handler{sessions: sessions, throttle: options.Throttle, keepalive: options.Keepalive}
	if h.throttle <= 0 {
		h.throttle = 250 * time.Millisecond
	}
	if h.keepalive <= 0 {
		h.keepalive = 15 * time.Second
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", h.state)
	mux.HandleFunc("GET /api/events", h.events)
	mux.HandleFunc("POST /api/sources", h.addSources)
	mux.HandleFunc("POST /api/entries/{id}/retry", h.retryEntry)
	mux.HandleFunc("POST /api/playlists/{index}/retry", h.retryPlaylist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reason := refusal(r); reason != "" {
			writeJSON(w, http.StatusForbidden, errorBody{Error: reason})
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// errorBody is every error the API returns.
type errorBody struct {
	Error string `json:"error"`
}

// writeJSON sends a JSON body. The run changes by the second, so nothing is cacheable.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
