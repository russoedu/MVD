package appwindow

import (
	"embed"
	"io/fs"
)

// The built page. The stage-web target copies it into web/ before Go compiles.
//
//go:embed all:web
var webFiles embed.FS

// pageAssets is the built page, which Wails serves itself to its web view: there is no
// server and no port.
func pageAssets() fs.FS {
	root, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	return root
}
