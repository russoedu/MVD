package install

import (
	"path/filepath"
	"strings"

	"youtube-downloader/apps/mvd/macbundle"
)

// installFolderName is the folder (or, on macOS, the .app bundle) the app lives in.
const installFolderName = "MVD"

// installKind says how the app is put in place, which differs by system.
type installKind int

const (
	// kindWindowsSystem is Program Files. Windows protects it, so it is only offered when
	// the app was started as an administrator, which is when a plain copy works.
	kindWindowsSystem installKind = iota
	// kindWindowsUser is the person's own Programs folder.
	kindWindowsUser
	// kindMacBundle is an MVD.app bundle, in /Applications or in ~/Applications.
	kindMacBundle
	// kindLinuxUser is ~/.local/bin, with an entry in the applications menu.
	kindLinuxUser
)

// installTarget is one place the app can be moved to.
type installTarget struct {
	Kind installKind
	// Everyone is true for a place that serves every account on the computer.
	Everyone bool
	// Folder is where it goes: a folder, or the .app bundle on macOS.
	Folder string
	// Program is the path of the program once it is there.
	Program string
}

// installPlaces are the folders of the machine that the targets are built from. Any may
// be empty, which leaves out the target that needs it. ProgramFiles is only filled in
// when the app may write there.
type installPlaces struct {
	ProgramFiles string
	LocalAppData string
	Home         string
}

// installTargets lists where the app belongs on goos, the place for everyone first and
// the person's own place after it.
//
//   - Windows: Program Files, and Programs under the local app data folder, where
//     per-user programs (VS Code, Obsidian, browsers) go.
//   - macOS: /Applications, and ~/Applications. Applications there are .app bundles, so
//     the app builds one around itself.
//   - Linux: ~/.local/bin only. There is no single applications folder, and a place for
//     everyone (such as /opt) would need root.
func installTargets(goos string, places installPlaces) []installTarget {
	var targets []installTarget

	switch goos {
	case "windows":
		if places.ProgramFiles != "" {
			folder := filepath.Join(places.ProgramFiles, installFolderName)
			targets = append(targets, installTarget{Kind: kindWindowsSystem, Everyone: true, Folder: folder, Program: filepath.Join(folder, "mvd.exe")})
		}
		if places.LocalAppData != "" {
			folder := filepath.Join(places.LocalAppData, "Programs", installFolderName)
			targets = append(targets, installTarget{Kind: kindWindowsUser, Folder: folder, Program: filepath.Join(folder, "mvd.exe")})
		}
	case "darwin":
		bundle := func(root string) installTarget {
			folder := filepath.Join(root, installFolderName+".app")

			return installTarget{Kind: kindMacBundle, Folder: folder, Program: macbundle.ProgramPath(folder)}
		}
		system := bundle("/Applications")
		system.Everyone = true
		targets = append(targets, system)
		if places.Home != "" {
			targets = append(targets, bundle(filepath.Join(places.Home, "Applications")))
		}
	case "linux":
		if places.Home != "" {
			folder := filepath.Join(places.Home, ".local", "bin")
			targets = append(targets, installTarget{Kind: kindLinuxUser, Folder: folder, Program: filepath.Join(folder, "mvd")})
		}
	}

	return targets
}

// samePlace reports whether two paths are the same. Windows paths are not case sensitive.
func samePlace(goos, a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if goos == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// isInstalled reports whether the program at exe is in one of the targets' folders.
func isInstalled(goos, exe string, targets []installTarget) bool {
	for _, target := range targets {
		if samePlace(goos, filepath.Dir(exe), filepath.Dir(target.Program)) {
			return true
		}
	}

	return false
}

// moveSituation is everything that decides whether to offer moving the app.
type moveSituation struct {
	// Version is "dev" for a build made on a developer's machine.
	Version string
	// Window is false when the app was asked to run without a window, which is how it
	// is run from a terminal or a script.
	Window bool
	// MovedFrom is set on the copy that a move has just started.
	MovedFrom string
	// Asked is whether the person has been offered the move before.
	Asked bool
	// HasTarget is false on a system with nowhere to move to.
	HasTarget bool
	// Installed is true when the program is already in one of the targets.
	Installed bool
}

// shouldOfferMove reports whether to ask the person to move the app. It asks once, the
// first time it is started from somewhere else, and never from a developer's build, a
// script, or the copy a move has just started.
func shouldOfferMove(s moveSituation) bool {
	switch {
	case !s.HasTarget:
		return false
	case s.Version == "dev":
		return false
	case !s.Window:
		return false
	case s.MovedFrom != "":
		return false
	case s.Asked:
		return false
	case s.Installed:
		return false
	}

	return true
}
