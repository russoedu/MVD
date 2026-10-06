package localserver

import (
	"encoding/json"
	"net/http"
)

// errorBody is every error the server returns.
type errorBody struct {
	Error string `json:"error"`
}

// writeJSON sends a JSON body that nothing may cache.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
