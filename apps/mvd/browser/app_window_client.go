package browser

import (
	"os"
	"os/exec"
	"runtime"
)

// OpenWindow shows url in a window of its own, like a desktop app, when a
// Chromium based browser is installed, and in the default browser otherwise. The
// url is always one this program built (http on loopback), never user input.
func OpenWindow(url string) error {
	h := host{
		goos: runtime.GOOS,
		env:  os.Getenv,
		exists: func(path string) bool {
			info, err := os.Stat(path)
			return err == nil && !info.IsDir()
		},
		lookPath: exec.LookPath,
	}
	for _, c := range appWindowCommands(h, url) {
		cmd := exec.Command(c.name, c.args...)
		if err := cmd.Start(); err != nil {
			continue
		}
		go func() { _ = cmd.Wait() }()
		return nil
	}
	return Open(url)
}
