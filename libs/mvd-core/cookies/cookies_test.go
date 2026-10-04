package cookies

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasYouTubeSession(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good.txt")
	os.WriteFile(good, []byte("# Netscape HTTP Cookie File\n#HttpOnly_.google.com\tTRUE\t/\tTRUE\t0\tSAPISID\tabc123\n"), 0600)
	if !hasYouTubeSession(good) {
		t.Error("should detect a Google session cookie")
	}

	bad := filepath.Join(dir, "bad.txt")
	os.WriteFile(bad, []byte("# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tFALSE\t0\tPREF\thl=en\n.example.com\tTRUE\t/\tTRUE\t0\tSID\tnope\n"), 0600)
	if hasYouTubeSession(bad) {
		t.Error("consent-only youtube cookies and a non-Google SID must not count")
	}

	empty := filepath.Join(dir, "empty.txt")
	os.WriteFile(empty, []byte("# Netscape HTTP Cookie File\n.google.com\tTRUE\t/\tTRUE\t0\tSID\t\n"), 0600)
	if hasYouTubeSession(empty) {
		t.Error("a session cookie with no value must not count")
	}

	if hasYouTubeSession(filepath.Join(dir, "missing.txt")) {
		t.Error("missing file should be false")
	}
}

func TestInstalledBrowsers(t *testing.T) {
	getenv := func(k string) string {
		switch k {
		case "APPDATA":
			return `C:\A`
		case "LOCALAPPDATA":
			return `C:\L`
		}
		return ""
	}
	// Build expected paths with the same joiner the policy uses, so the test
	// is separator-agnostic across the CI matrix.
	chrome := filepath.Join(`C:\L`, "Google", "Chrome", "User Data")
	firefox := filepath.Join(`C:\A`, "Mozilla", "Firefox", "Profiles")
	present := map[string]bool{chrome: true, firefox: true}
	exists := func(p string) bool { return present[p] }

	got := installedBrowsers("windows", getenv, "", exists)
	if len(got) != 2 || got[0] != "firefox" || got[1] != "chrome" {
		t.Errorf("want [firefox chrome], got %v", got)
	}

	if none := installedBrowsers("windows", getenv, "", func(string) bool { return false }); len(none) != 0 {
		t.Errorf("nothing installed should be empty, got %v", none)
	}
}
