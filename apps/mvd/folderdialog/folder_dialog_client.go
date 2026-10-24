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
)

// Dialog shows the operating system's own folder chooser, on the screen of the person
// at this machine.
type Dialog struct{}

// Pick runs the chooser and waits. Cancelling the context (the page went away)
// closes it.
func (Dialog) Pick(ctx context.Context, start string) (string, bool, error) {
	if hasNativeDialog {
		return nativeDialog(ctx, dialogRequest{kind: pickFolder, start: start})
	}
	command, ok := folderDialogCommand(runtime.GOOS, start, oscommand.HasProgram)
	if !ok {
		return "", false, ErrNoDialog
	}

	return run(ctx, command)
}

// PickFile runs the system's open-file window and waits. extension (without its dot)
// and label describe the files it shows first, such as "mvd" and "MVD review".
func (Dialog) PickFile(ctx context.Context, start, extension, label string) (string, bool, error) {
	if hasNativeDialog {
		return nativeDialog(ctx, dialogRequest{kind: openFile, start: start, extension: extension, label: label})
	}
	command, ok := fileDialogCommand(runtime.GOOS, start, extension, label, oscommand.HasProgram)
	if !ok {
		return "", false, ErrNoDialog
	}

	return run(ctx, command)
}

// SaveFile runs the system's save-file window and waits. start is the folder it opens
// in and name the file name offered; extension (without its dot) and label describe the
// file type, and the extension is added when the person typed a name without one.
func (Dialog) SaveFile(ctx context.Context, start, name, extension, label string) (string, bool, error) {
	if hasNativeDialog {
		path, chosen, err := nativeDialog(ctx, dialogRequest{kind: saveFile, start: start, name: name, extension: extension, label: label})

		return withExtension(path, extension), chosen, err
	}
	command, ok := saveDialogCommand(runtime.GOOS, start, name, extension, label, oscommand.HasProgram)
	if !ok {
		return "", false, ErrNoDialog
	}
	path, chosen, err := run(ctx, command)

	return withExtension(path, extension), chosen, err
}

// run starts a chooser program and reads what the person chose.
func run(ctx context.Context, command oscommand.Command) (string, bool, error) {
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
