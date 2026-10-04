// Package cookies acquires a usable YouTube cookie file by trying the
// browsers installed on this machine. It depends only on the ytdlp slice.
package cookies

import (
	"os"
	"path/filepath"
	"runtime"
)

// browserCand is a yt-dlp browser spec and the profile directories whose
// presence means that browser is installed.
type browserCand struct {
	spec string
	dirs []string
}

// candidates lists browsers to try, most reliable first. Firefox leads
// because Chrome/Edge App-Bound Encryption (Chrome/Edge 127+) often stops
// yt-dlp from decrypting their cookies at all.
func candidates(goos string, getenv func(string) string, home string) []browserCand {
	join := filepath.Join
	switch goos {
	case "windows":
		appData := getenv("APPDATA")
		local := getenv("LOCALAPPDATA")
		return []browserCand{
			{"firefox", []string{join(appData, "Mozilla", "Firefox", "Profiles")}},
			{"edge", []string{join(local, "Microsoft", "Edge", "User Data")}},
			{"chrome", []string{join(local, "Google", "Chrome", "User Data")}},
			{"brave", []string{join(local, "BraveSoftware", "Brave-Browser", "User Data")}},
			{"chromium", []string{join(local, "Chromium", "User Data")}},
			{"opera", []string{join(appData, "Opera Software", "Opera Stable")}},
			{"vivaldi", []string{join(local, "Vivaldi", "User Data")}},
		}
	case "darwin":
		as := join(home, "Library", "Application Support")
		return []browserCand{
			{"firefox", []string{join(as, "Firefox", "Profiles")}},
			{"edge", []string{join(as, "Microsoft Edge")}},
			{"chrome", []string{join(as, "Google", "Chrome")}},
			{"brave", []string{join(as, "BraveSoftware", "Brave-Browser")}},
			{"chromium", []string{join(as, "Chromium")}},
			{"opera", []string{join(as, "com.operasoftware.Opera")}},
			{"vivaldi", []string{join(as, "Vivaldi")}},
		}
	default: // linux and the rest
		cfg := join(home, ".config")
		return []browserCand{
			{"firefox", []string{join(home, ".mozilla", "firefox"), join(home, "snap", "firefox", "common", ".mozilla", "firefox")}},
			{"edge", []string{join(cfg, "microsoft-edge")}},
			{"chrome", []string{join(cfg, "google-chrome")}},
			{"brave", []string{join(cfg, "BraveSoftware", "Brave-Browser")}},
			{"chromium", []string{join(cfg, "chromium")}},
			{"opera", []string{join(cfg, "opera")}},
			{"vivaldi", []string{join(cfg, "vivaldi")}},
		}
	}
}

// installedBrowsers returns the specs whose profile directory exists.
func installedBrowsers(goos string, getenv func(string) string, home string, exists func(string) bool) []string {
	var out []string
	for _, b := range candidates(goos, getenv, home) {
		for _, d := range b.dirs {
			if exists(d) {
				out = append(out, b.spec)
				break
			}
		}
	}
	return out
}

// InstalledBrowsers detects the browsers on this machine, most reliable first.
func InstalledBrowsers() []string {
	home, _ := os.UserHomeDir()
	return installedBrowsers(runtime.GOOS, os.Getenv, home, dirExists)
}

func dirExists(p string) bool {
	if p == "" {
		return false
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
