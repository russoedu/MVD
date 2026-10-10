// Package openurl opens a web address in the person's own browser.
package openurl

import (
	"errors"
	"net/url"
	"os/exec"
	"runtime"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// Open shows address in the default browser. Only http and https addresses are opened: the
// address can come from a file a person was given, and must not start anything else.
func Open(address string) error {
	u, err := url.Parse(address)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("only web addresses can be opened")
	}
	name, args := command(runtime.GOOS, address)
	cmd := exec.Command(name, args...)
	procwindow.Hide(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	// The browser is not waited for, and it is not ours to stop.
	go func() { _ = cmd.Wait() }()
	return nil
}

// command is the program that opens an address on an operating system.
func command(goos, address string) (name string, args []string) {
	switch goos {
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", address}
	case "darwin":
		return "open", []string{address}
	}
	return "xdg-open", []string{address}
}
