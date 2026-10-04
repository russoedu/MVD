package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"youtube-downloader/libs/mvd-core/procwindow"
)

// offerMoveHere is offerMove against the real machine. It reports true when the app has
// been moved and started from its new place, and the caller should exit.
func offerMoveHere(appDir string, tray bool, movedFrom string) bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	return offerMove(moveEnvironment{
		GOOS: runtime.GOOS, Version: version, Tray: tray, MovedFrom: movedFrom,
		LocalAppData: os.Getenv("LOCALAPPDATA"),
		StartMenu:    filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"),
		AppDir:       appDir,
		Exe:          exe,
		Args:         os.Args[1:],
		Ask:          askYesNo,
		Tell:         showFatal,
		Copy:         copyExecutable,
		Shortcut:     makeShortcut,
		Start:        startProgram,
	})
}

// cleanUpMovedProgram removes the copy a move left behind, once it has exited. It never
// removes the program that is running it.
func cleanUpMovedProgram(path string) {
	if exe, err := os.Executable(); err == nil && samePlace(runtime.GOOS, exe, path) {
		return
	}
	removeMovedProgram(path)
}

// makeShortcut creates a shortcut file with PowerShell, which is how Windows lets a
// program make one without extra libraries.
func makeShortcut(link, target string) error {
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	command := shortcutCommand(link, target)
	cmd := exec.Command(command.name, command.args...)
	cmd.Env = append(os.Environ(), command.env...)
	procwindow.Hide(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}

	return nil
}

// startProgram starts path, in its own folder, and does not wait for it.
func startProgram(path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.Dir = filepath.Dir(path)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()

	return nil
}
