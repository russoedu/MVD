package songfile

import (
	"path/filepath"
	"strings"
)

// FilePath returns the path a download-list entry points at when it names a
// song file: no URL scheme and a .txt, .csv or .tsv extension. Quotes around
// the path (what "Copy as path" gives) are dropped.
func FilePath(link string) (string, bool) {
	path := strings.Trim(strings.TrimSpace(link), `"'`)
	if path == "" || strings.Contains(path, "://") {
		return "", false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".csv", ".tsv":
		return path, true
	}
	return "", false
}
