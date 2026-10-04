package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// movedMarkerName is the file in the app-data folder that records the person has been
// asked to move the app, whatever they answered.
const movedMarkerName = "install-offered"

// installedProgramName is the file name the app has once it has been moved, whatever it
// was called before (a browser may have saved it as "mvd-tray (1).exe").
func installedProgramName(goos string) string {
	if goos == "windows" {
		return "mvd-tray.exe"
	}

	return "mvd-tray"
}

// moveEnvironment is what moving the app needs from the machine, so that every branch
// can be tested without a screen or a real installation.
type moveEnvironment struct {
	GOOS      string
	Version   string
	Tray      bool
	MovedFrom string
	// LocalAppData holds the folder the app is moved into; StartMenu is where its
	// shortcut goes; AppDir is where the "already asked" marker is kept.
	LocalAppData string
	StartMenu    string
	AppDir       string
	// Exe is the running program and Args the arguments it was started with.
	Exe  string
	Args []string

	// Ask puts a yes or no question to the person.
	Ask func(title, text string) bool
	// Tell shows the person a problem.
	Tell func(text string)
	// Copy copies the program to its new place.
	Copy func(src, dst string) error
	// Shortcut makes a shortcut to the program.
	Shortcut func(link, target string) error
	// Start starts the moved program.
	Start func(path string, args []string) error
}

// offerMove asks the person, the first time the app is started from somewhere it does
// not belong, whether to move it there. If they agree it copies itself, adds a shortcut
// to the Start menu, starts the copy and returns true, and the caller should exit: the
// copy takes over, and removes this one once it has gone (see removeAfterExit).
//
// The question is recorded as asked before it is put, so whatever the answer, and
// however the app ends, it is not asked again. If anything goes wrong the app keeps
// running from where it is and says why.
func offerMove(env moveEnvironment) (moved bool) {
	target := installFolder(env.GOOS, env.LocalAppData)
	marker := filepath.Join(env.AppDir, movedMarkerName)
	_, markerErr := os.Stat(marker)

	if !shouldOfferMove(moveSituation{
		GOOS: env.GOOS, Version: env.Version, Tray: env.Tray, MovedFrom: env.MovedFrom,
		Asked: markerErr == nil, ExeFolder: filepath.Dir(env.Exe), Target: target,
	}) {
		return false
	}
	if err := os.WriteFile(marker, []byte("asked\n"), 0o600); err != nil {
		// Not being able to remember would mean asking at every start.
		return false
	}

	question := fmt.Sprintf(
		"MVD is running from:\n\n%s\n\nMove it to its own folder, so it can be started from the Start menu?\n\n%s\n\n"+
			"Yes moves it and starts it from there. No leaves it where it is, and you will not be asked again.",
		filepath.Dir(env.Exe), target)
	if !env.Ask("MVD", question) {
		return false
	}

	destination := filepath.Join(target, installedProgramName(env.GOOS))
	if err := env.Copy(env.Exe, destination); err != nil {
		env.Tell(fmt.Sprintf("MVD could not move itself to %s: %v\n\nIt will keep running from where it is.", target, err))

		return false
	}
	// A missing shortcut is a loss of convenience, not a reason to undo the move.
	_ = env.Shortcut(filepath.Join(env.StartMenu, "MVD.lnk"), destination)

	args := append(append([]string{}, env.Args...), "-moved-from="+env.Exe)
	if err := env.Start(destination, args); err != nil {
		env.Tell(fmt.Sprintf("MVD was copied to %s but could not be started from there: %v\n\nIt will keep running from where it is.", target, err))

		return false
	}

	return true
}

// copyExecutable copies the program at src to dst, creating dst's folder. It writes to a
// temporary name first, so a copy that is interrupted never leaves half a program where
// the Start menu points. An existing program at dst is replaced, which fails, saying so,
// if it is running.
func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	partial := dst + ".part"
	out, err := os.OpenFile(partial, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(partial)

		return err
	}

	if _, statErr := os.Stat(dst); statErr == nil {
		if err := os.Remove(dst); err != nil {
			_ = os.Remove(partial)

			return fmt.Errorf("the copy already there cannot be replaced, probably because it is running: %w", err)
		}
	}
	if err := os.Rename(partial, dst); err != nil {
		_ = os.Remove(partial)

		return err
	}

	return nil
}
