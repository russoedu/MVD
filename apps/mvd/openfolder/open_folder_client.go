// Package openfolder shows a folder in the system's file manager.
package openfolder

import (
	"errors"
	"os"
	"os/exec"
	"runtime"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// Open shows the folder at path in the file manager of the system (Explorer, Finder, or what
// xdg-open picks). The file manager is not waited for, and it is not ours to stop.
func Open(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("not a folder: " + path)
	}
	name, args := command(runtime.GOOS, path)
	cmd := exec.Command(name, args...)
	procwindow.Hide(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()

	return nil
}

// command is the program that shows a folder on an operating system.
func command(goos, path string) (name string, args []string) {
	switch goos {
	case "windows":
		return "explorer.exe", []string{path}
	case "darwin":
		return "open", []string{path}
	}

	return "xdg-open", []string{path}
}
