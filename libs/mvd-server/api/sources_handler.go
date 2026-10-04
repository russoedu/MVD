package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"youtube-downloader/libs/mvd-server/session"
)

// maxSourcesBody is far more than any list a person pastes, and keeps a runaway
// client from filling memory.
const maxSourcesBody = 1 << 20

// sourcesRequest is what the browser sends: the lines as a list, or as the block of
// text a textarea gives (one URL per line). Either is fine; both are added.
type sourcesRequest struct {
	URLs []string `json:"urls"`
	Text string   `json:"text"`
}

// addSources queues what the browser sent and says what became of each line.
func (h *handler) addSources(w http.ResponseWriter, r *http.Request) {
	var body sourcesRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSourcesBody))
	// Unknown fields are an error, so a misspelt key is not silently an empty paste.
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "expected {\"urls\": [...]} or {\"text\": \"...\"}"})
		return
	}

	lines := append([]string{}, body.URLs...)
	if body.Text != "" {
		lines = append(lines, strings.Split(strings.ReplaceAll(body.Text, "\r\n", "\n"), "\n")...)
	}

	result, err := h.sessions.Add(lines)
	switch {
	case errors.Is(err, session.ErrClosed):
		writeJSON(w, http.StatusServiceUnavailable, errorBody{Error: "MVD is shutting down"})
	case err != nil:
		// Not the user's mistake: yt-dlp is missing, or cookies could not be read.
		writeJSON(w, http.StatusBadGateway, errorBody{Error: err.Error()})
	default:
		writeJSON(w, http.StatusOK, result)
	}
}
