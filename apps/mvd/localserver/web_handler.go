package localserver

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

// The built frontend. The stage-web target copies it into web/ before Go compiles.
//
//go:embed all:web
var webFiles embed.FS

// webHandler serves the built frontend. A path that is not a file falls back to
// index.html, so a client-side route survives a reload.
func webHandler() http.Handler {
	root, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := path.Clean(r.URL.Path)[1:]
		if name != "" {
			if _, err := fs.Stat(root, name); err != nil {
				r.URL.Path = "/"
			}
		}
		files.ServeHTTP(w, r)
	})
}
