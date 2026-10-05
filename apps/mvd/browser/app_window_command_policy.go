package browser

import "path/filepath"

// command is a program to start and its arguments.
type command struct {
	name string
	args []string
}

// host is what the policy needs to know about the machine.
type host struct {
	goos     string
	env      func(name string) string
	exists   func(path string) bool
	lookPath func(name string) (string, error)
}

const windowSize = "--window-size=1100,760"

// appWindowCommands lists, best first, the commands that show url in a window of
// its own, without tabs or an address bar: a Chromium based browser's app mode.
// Edge is first on Windows because it is always there. It is empty when no such
// browser is installed; the caller then uses the default browser.
func appWindowCommands(h host, url string) []command {
	var cmds []command
	add := func(path string) {
		cmds = append(cmds, command{name: path, args: []string{"--app=" + url, windowSize}})
	}

	switch h.goos {
	case "windows":
		for _, dir := range []string{h.env("ProgramFiles(x86)"), h.env("ProgramFiles"), h.env("LocalAppData")} {
			if dir == "" {
				continue
			}
			for _, rel := range []string{
				filepath.Join("Microsoft", "Edge", "Application", "msedge.exe"),
				filepath.Join("Google", "Chrome", "Application", "chrome.exe"),
			} {
				if path := filepath.Join(dir, rel); h.exists(path) {
					add(path)
				}
			}
		}
	case "darwin":
		for _, path := range []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		} {
			if h.exists(path) {
				add(path)
			}
		}
	default:
		for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge"} {
			if path, err := h.lookPath(name); err == nil {
				add(path)
			}
		}
	}
	return cmds
}
