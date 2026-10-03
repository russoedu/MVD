// Package appdir locates the per-user folders MVD stores its files in: the
// app-data/preferences directory for config and the saved list, and the OS
// Downloads directory used as the default output.
package appdir

import (
	"os"
	"path/filepath"
)

// Dir returns the MVD app-data directory (created if missing), for example
// %AppData%\mvd on Windows, ~/Library/Application Support/mvd on macOS or
// ~/.config/mvd on Linux. It falls back to ./.mvd when the OS config dir
// cannot be determined.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = ".mvd-fallback"
	}
	dir := filepath.Join(base, "mvd")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// DefaultDownloadsDir returns the user's Downloads folder, or the home
// directory when it cannot be found.
func DefaultDownloadsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	downloads := filepath.Join(home, "Downloads")
	if info, err := os.Stat(downloads); err == nil && info.IsDir() {
		return downloads
	}
	return home
}
