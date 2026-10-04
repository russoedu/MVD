package folderdialog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
	"youtube-downloader/libs/mvd-core/procwindow"
	"youtube-downloader/libs/mvd-server/api"
)

// Dialog shows the operating system's own folder chooser. It is the
// api.FolderPicker of the real app: the page asks, the person answers on their screen.
type Dialog struct{}

// Pick runs the chooser and waits. Cancelling the context (the page went away)
// closes it.
func (Dialog) Pick(ctx context.Context, start string) (string, bool, error) {
	command, ok := folderDialogCommand(runtime.GOOS, start, oscommand.HasProgram)
	if !ok {
		return "", false, api.ErrNoDialog
	}

	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	procwindow.Hide(cmd)
	cmd.Env = append(os.Environ(), command.Env...)
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
			return "", false, fmt.Errorf("%s: %s", command.Name, message)
		}

		return "", false, fmt.Errorf("%s: %w", command.Name, err)
	}

	path, chosen := folderChoice(stdout.String())

	return path, chosen, nil
}
