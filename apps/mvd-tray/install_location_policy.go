package main

import (
	"path/filepath"
	"strings"
)

// installFolderName is the folder the app lives in once it has been moved.
const installFolderName = "MVD"

// installFolder is the folder the app belongs in on goos, or "" where it does not
// manage its own location.
//
// On Windows that is Programs under the person's local app data, which is where
// per-user programs (VS Code, Obsidian, browsers) go: it needs no administrator rights
// and can always be written to, unlike Program Files. macOS applications are .app
// bundles and this app ships as a bare program, so it has no place to move to there,
// and Linux has no single applications folder.
func installFolder(goos, localAppData string) string {
	if goos != "windows" || strings.TrimSpace(localAppData) == "" {
		return ""
	}

	return filepath.Join(localAppData, "Programs", installFolderName)
}

// samePlace reports whether two paths are the same folder. Windows paths are not case
// sensitive.
func samePlace(goos, a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if goos == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// moveSituation is everything that decides whether to offer moving the app.
type moveSituation struct {
	GOOS string
	// Version is "dev" for a build made on a developer's machine.
	Version string
	// Tray is false when the app was asked to run without a tray icon, which is how it
	// is run from a terminal or a script.
	Tray bool
	// MovedFrom is set on the copy that a move has just started.
	MovedFrom string
	// Asked is whether the person has been offered the move before.
	Asked bool
	// ExeFolder is the folder the running program is in.
	ExeFolder string
	// Target is the folder it belongs in, or "" if it has none on this OS.
	Target string
}

// shouldOfferMove reports whether to ask the person to move the app. It asks once, the
// first time it is started from somewhere else, and never from a developer's build, a
// script, or the copy a move has just started.
func shouldOfferMove(s moveSituation) bool {
	switch {
	case s.Target == "":
		return false
	case s.Version == "dev":
		return false
	case !s.Tray:
		return false
	case s.MovedFrom != "":
		return false
	case s.Asked:
		return false
	case samePlace(s.GOOS, s.ExeFolder, s.Target):
		return false
	}

	return true
}

// shortcutScript creates a shortcut. The paths come from the environment, so a folder
// name can never be read as code.
const shortcutScript = `$ErrorActionPreference = 'Stop'
$shortcut = (New-Object -ComObject WScript.Shell).CreateShortcut($env:MVD_SHORTCUT)
$shortcut.TargetPath = $env:MVD_TARGET
$shortcut.WorkingDirectory = $env:MVD_FOLDER
$shortcut.Description = 'MVD, the music video downloader'
$shortcut.Save()
`

// shortcutCommand returns the command that makes a shortcut at link to target.
func shortcutCommand(link, target string) dialogCommand {
	return dialogCommand{
		name: "powershell",
		args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShell(shortcutScript)},
		env: []string{
			"MVD_SHORTCUT=" + link,
			"MVD_TARGET=" + target,
			"MVD_FOLDER=" + filepath.Dir(target),
		},
	}
}
