package main

import (
	"os"
	"os/exec"
	"runtime"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// notifyUser shows n through the operating system and returns at once. It is best
// effort: where nothing can show a notification, nothing happens.
func notifyUser(n userNotice) {
	command, ok := notificationCommand(runtime.GOOS, n, hasProgram)
	if !ok {
		return
	}

	cmd := exec.Command(command.name, command.args...)
	cmd.Env = append(os.Environ(), command.env...)
	procwindow.Hide(cmd)
	if err := cmd.Start(); err != nil {
		return
	}
	go func() { _ = cmd.Wait() }()
}
