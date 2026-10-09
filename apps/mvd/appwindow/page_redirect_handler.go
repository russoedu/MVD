package appwindow

import (
	"encoding/json"
	"net/http"
)

// pageHandler is what the window loads first: a blank dark page that sends the web view on
// to the app's own address. The page itself, and the terminal's WebSocket, are served by the
// local server, so the window sees exactly what a browser at that address would, and the
// WebSocket stays on the same origin as the page.
func pageHandler(url string) http.Handler {
	target, _ := json.Marshal(url) // a quoted, escaped JavaScript string
	page := `<!doctype html>
<html><head><meta charset="utf-8"><title>MVD</title>
<style>html,body{margin:0;height:100%;background:#000;color:#9a9a9a;font:14px monospace}</style>
</head><body><p style="padding:1em">Opening MVD...</p>
<script>location.replace(` + string(target) + `)</script></body></html>`

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(page))
	})
}
