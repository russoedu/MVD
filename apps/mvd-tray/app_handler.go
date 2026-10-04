package main

import (
	"net/http"

	"youtube-downloader/libs/mvd-server/api"
)

// newAppHandler is the whole site: the API under /api/ and the built frontend for
// everything else.
func newAppHandler(sessions api.Sessions) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", api.New(sessions, api.Options{}))
	mux.Handle("/", webHandler())
	return mux
}
