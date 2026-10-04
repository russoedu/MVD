package notification

import (
	"os"
	"os/exec"
	"runtime"

	"youtube-downloader/apps/mvd-tray/oscommand"
	"youtube-downloader/libs/mvd-core/procwindow"
)

// Notify shows n through the operating system and returns at once. It is best
// effort: where nothing can show a notification, nothing happens.
func Notify(n Notice) {
	command, ok := notificationCommand(runtime.GOOS, n, oscommand.HasProgram)
	if !ok {
		return
	}

	cmd := exec.Command(command.Name, command.Args...)
	cmd.Env = append(os.Environ(), command.Env...)
	procwindow.Hide(cmd)
	if err := cmd.Start(); err != nil {
		return
	}
	go func() { _ = cmd.Wait() }()
}
