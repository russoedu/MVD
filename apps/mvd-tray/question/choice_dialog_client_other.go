//go:build !windows

package question

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"

	"youtube-downloader/apps/mvd-tray/oscommand"
)

// Ask puts the question to the person with the system's own dialog (osascript on
// macOS, zenity or kdialog on Linux). With no way to show one it says so instead of
// guessing an answer.
func Ask(title, text string, choices []string) Answer {
	command, ok := questionCommand(runtime.GOOS, title, text, choices, oscommand.HasProgram)
	if !ok {
		return AnswerUnavailable
	}

	cmd := exec.Command(command.Name, command.Args...)
	cmd.Env = append(os.Environ(), command.Env...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return AnswerUnavailable
		}
		code = exit.ExitCode()
	}

	return parseAnswer(runtime.GOOS, choices, code, stdout.String())
}
