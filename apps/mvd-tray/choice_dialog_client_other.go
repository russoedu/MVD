//go:build !windows

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"
)

// askChoice puts the question to the person with the system's own dialog (osascript on
// macOS, zenity or kdialog on Linux). With no way to show one it says so instead of
// guessing an answer.
func askChoice(title, text string, choices []string) answer {
	command, ok := questionCommand(runtime.GOOS, title, text, choices, hasProgram)
	if !ok {
		return answerUnavailable
	}

	cmd := exec.Command(command.name, command.args...)
	cmd.Env = append(os.Environ(), command.env...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return answerUnavailable
		}
		code = exit.ExitCode()
	}

	return parseAnswer(runtime.GOOS, choices, code, stdout.String())
}
