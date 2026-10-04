package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"youtube-downloader/libs/mvd-core/procwindow"
	"youtube-downloader/libs/mvd-server/api"
)

// folderDialog shows the operating system's own folder chooser. It is the
// api.FolderPicker of the real app: the page asks, the person answers on their screen.
type folderDialog struct{}

// Pick runs the chooser and waits. Cancelling the context (the page went away)
// closes it.
func (folderDialog) Pick(ctx context.Context, start string) (string, bool, error) {
	command, ok := folderDialogCommand(runtime.GOOS, start, hasProgram)
	if !ok {
		return "", false, api.ErrNoDialog
	}

	cmd := exec.CommandContext(ctx, command.name, command.args...)
	procwindow.Hide(cmd)
	cmd.Env = append(os.Environ(), command.env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && cancelledByExit(runtime.GOOS, exit.ExitCode(), stderr.String()) {
			return "", false, nil
		}
		if ctx.Err() != nil {
			return "", false, ctx.Err()
		}
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", false, fmt.Errorf("%s: %s", command.name, message)
		}

		return "", false, fmt.Errorf("%s: %w", command.name, err)
	}

	path, chosen := folderChoice(stdout.String())

	return path, chosen, nil
}
