package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"youtube-downloader/apps/mvd/console"
	"youtube-downloader/apps/mvd/question"
	"youtube-downloader/libs/mvd-core/procwindow"
)

// MoveResult is how a move that the person asked for went.
type MoveResult int

const (
	// Moved: the app was installed and the copy started; the caller should quit.
	Moved MoveResult = iota
	// Declined: the person said no, or the move failed and said so.
	Declined
	// AlreadyThere: the app already lives in its own folder.
	AlreadyThere
	// CannotMove: there is nowhere to move to on this machine, or no way to ask.
	CannotMove
)

// OfferMoveHere is the offer to move the app, made by itself the first time it is started
// from somewhere it does not belong, against the real machine. It reports true when the
// app has been moved and started from its new place, and the caller should exit.
func OfferMoveHere(appDir string, window bool, movedFrom, version string) bool {
	env, ok := realEnvironment(appDir, window, movedFrom, version)
	if !ok {
		return false
	}

	return offerMove(env)
}

// MoveNow is the move the person asked for, from the preferences or with -move: they
// are asked whatever they answered before, and nothing is remembered.
func MoveNow(appDir, version string) MoveResult {
	env, ok := realEnvironment(appDir, true, "", version)
	if !ok {
		return CannotMove
	}
	env.Force = true

	switch moveOffer(env) {
	case offerMoved:
		return Moved
	case offerAlreadyThere:
		return AlreadyThere
	case offerCannot, offerSkipped:
		return CannotMove
	}

	return Declined
}

// realEnvironment describes this machine and this run to the move.
func realEnvironment(appDir string, window bool, movedFrom, version string) (moveEnvironment, bool) {
	exe, err := os.Executable()
	if err != nil {
		return moveEnvironment{}, false
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	home, _ := os.UserHomeDir()
	// Program Files can only be written to by an administrator, so it is only a place to
	// move to when the app is running as one. The app never asks for those rights itself:
	// an unsigned program that can start an administrator step is treated as malware by
	// Windows' antivirus, which then removes it.
	programFiles := ""
	if runtime.GOOS == "windows" && isElevated() {
		programFiles = os.Getenv("ProgramFiles")
	}

	return moveEnvironment{
		GOOS: runtime.GOOS, Version: version, Window: window, MovedFrom: movedFrom,
		Places:    installPlaces{ProgramFiles: programFiles, LocalAppData: os.Getenv("LOCALAPPDATA"), Home: home},
		AdminHint: runtime.GOOS == "windows" && !isElevated(),
		AppDir:    appDir,
		Exe:       exe,
		Args:      os.Args[1:],
		Ask:       question.Ask,
		Tell:      console.ShowFatal,
		Install: func(target installTarget, exe string) error {
			return installOnThisMachine(target, exe, version)
		},
		Start: startInstalled,
	}, true
}

// installOnThisMachine puts the program in target in the way that system needs.
func installOnThisMachine(target installTarget, exe, version string) error {
	switch target.Kind {
	case kindWindowsSystem:
		return installWindowsSystem(target, exe, allUsersStartMenuLink(), makeShortcut, func(t installTarget) error { return registerUninstallEntry(t, version) })
	case kindWindowsUser:
		return installWindowsUser(target, exe, userStartMenu(), makeShortcut, func(t installTarget) error { return registerUninstallEntry(t, version) })
	case kindMacBundle:
		return installMacBundle(target, exe, version)
	case kindLinuxUser:
		return installLinuxUser(target, exe, userApplicationsMenu())
	}

	return errors.New("this kind of place is not supported")
}

// FootprintHere is what the installer can have put on this machine.
func FootprintHere() Footprint {
	home, _ := os.UserHomeDir()
	places := installPlaces{ProgramFiles: os.Getenv("ProgramFiles"), LocalAppData: os.Getenv("LOCALAPPDATA"), Home: home}

	return footprintOf(runtime.GOOS, places, userStartMenu(), allUsersStartMenuLink(), userApplicationsMenu())
}

// allUsersStartMenuLink is where the Start menu shortcut for every account goes.
func allUsersStartMenuLink() string {
	return filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs", "MVD.lnk")
}

// userStartMenu is the person's own Start menu folder.
func userStartMenu() string {
	return filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
}

// userApplicationsMenu is where a Linux desktop looks for the person's own menu entries.
func userApplicationsMenu() string {
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, _ := os.UserHomeDir()
		data = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(data, "applications")
}

// CleanUpMovedProgram removes the copy a move left behind, once it has exited. It never
// removes the program that is running it.
func CleanUpMovedProgram(path string) {
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
	cmd := exec.Command(command.Name, command.Args...)
	cmd.Env = append(os.Environ(), command.Env...)
	procwindow.Hide(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}

	return nil
}

// startInstalled starts the installed program, in its own folder, and does not wait for
// it. On macOS that goes through the system, which is what applies the bundle's settings.
func startInstalled(target installTarget, args []string) error {
	var cmd *exec.Cmd
	if target.Kind == kindMacBundle {
		cmd = exec.Command("open", append([]string{"-n", target.Folder, "--args"}, args...)...)
	} else {
		cmd = exec.Command(target.Program, args...)
		cmd.Dir = filepath.Dir(target.Program)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()

	return nil
}
