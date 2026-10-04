package main

import (
	"errors"
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
	home, _ := os.UserHomeDir()

	return offerMove(moveEnvironment{
		GOOS: runtime.GOOS, Version: version, Tray: tray, MovedFrom: movedFrom,
		Places:  installPlaces{ProgramFiles: os.Getenv("ProgramFiles"), LocalAppData: os.Getenv("LOCALAPPDATA"), Home: home},
		AppDir:  appDir,
		Exe:     exe,
		Args:    os.Args[1:],
		Ask:     askChoice,
		Tell:    showFatal,
		Install: installOnThisMachine,
		Start:   startInstalled,
	})
}

// installOnThisMachine puts the program in target in the way that system needs.
func installOnThisMachine(target installTarget, exe string) error {
	switch target.Kind {
	case kindWindowsSystem:
		return elevatedInstall(exe, target.Program, allUsersStartMenuLink())
	case kindWindowsUser:
		return installWindowsUser(target, exe, userStartMenu(), makeShortcut)
	case kindMacBundle:
		return installMacBundle(target, exe, version)
	case kindLinuxUser:
		return installLinuxUser(target, exe, userApplicationsMenu())
	}

	return errors.New("this kind of place is not supported")
}

// allUsersStartMenuLink is where the Start menu shortcut for every account goes.
func allUsersStartMenuLink() string {
	return filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs", "MVD.lnk")
}

// elevatedInstall installs the program from source to dest with administrator approval,
// which Windows asks the person for. It returns an error saying so if they decline.
func elevatedInstall(source, dest, link string) error {
	command := elevationCommand(source, dest, link)
	cmd := exec.Command(command.name, command.args...)
	procwindow.Hide(cmd)

	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &exit) && exit.ExitCode() == exitDeclined:
		return errors.New("administrator approval was declined")
	case errors.As(err, &exit):
		return fmt.Errorf("the administrator step failed (exit code %d)", exit.ExitCode())
	}

	return err
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
