package browser

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func installed(paths ...string) func(string) bool {
	return func(path string) bool {
		for _, p := range paths {
			if p == path {
				return true
			}
		}
		return false
	}
}

func TestWindowsPrefersEdgeAndFallsBackToChrome(t *testing.T) {
	programFiles := filepath.Join("C:", "Program Files")
	edge := filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe")
	chrome := filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe")
	env := func(name string) string {
		if name == "ProgramFiles" {
			return programFiles
		}
		return ""
	}

	both := appWindowCommands(host{goos: "windows", env: env, exists: installed(edge, chrome)}, "http://127.0.0.1:8421/terminal")
	want := []command{
		{edge, []string{"--app=http://127.0.0.1:8421/terminal", windowSize}},
		{chrome, []string{"--app=http://127.0.0.1:8421/terminal", windowSize}},
	}
	if !reflect.DeepEqual(both, want) {
		t.Errorf("got %v, want %v", both, want)
	}

	onlyChrome := appWindowCommands(host{goos: "windows", env: env, exists: installed(chrome)}, "http://x")
	if len(onlyChrome) != 1 || onlyChrome[0].name != chrome {
		t.Errorf("only Chrome is installed, got %v", onlyChrome)
	}
}

func TestNoChromiumBrowserMeansNoCommands(t *testing.T) {
	none := host{
		goos:     "linux",
		env:      func(string) string { return "" },
		exists:   installed(),
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
	}
	if got := appWindowCommands(none, "http://x"); len(got) != 0 {
		t.Errorf("want none, got %v", got)
	}
	none.goos = "darwin"
	if got := appWindowCommands(none, "http://x"); len(got) != 0 {
		t.Errorf("want none on macOS, got %v", got)
	}
}

func TestLinuxLooksForChromiumBrowsersOnThePath(t *testing.T) {
	h := host{
		goos: "linux",
		env:  func(string) string { return "" },
		lookPath: func(name string) (string, error) {
			if name == "chromium" {
				return "/usr/bin/chromium", nil
			}
			return "", errors.New("not found")
		},
	}
	got := appWindowCommands(h, "http://x")
	if len(got) != 1 || got[0].name != "/usr/bin/chromium" || got[0].args[0] != "--app=http://x" {
		t.Errorf("got %v", got)
	}
}

func TestMacOSFindsAnInstalledAppBundle(t *testing.T) {
	chrome := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	got := appWindowCommands(host{goos: "darwin", env: func(string) string { return "" }, exists: installed(chrome)}, "http://x")
	if len(got) != 1 || got[0].name != chrome {
		t.Errorf("got %v", got)
	}
}
