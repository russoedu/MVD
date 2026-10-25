package folderbrowser

import (
	"os"
	"path/filepath"
)

// ResolveDir is the absolute folder the browser begins in: start, or the working folder when
// start is empty.
func ResolveDir(start string) string {
	if start == "" {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}

		return "."
	}
	if abs, err := filepath.Abs(start); err == nil {
		return abs
	}

	return start
}
