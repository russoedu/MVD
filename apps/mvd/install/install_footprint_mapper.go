package install

import "path/filepath"

// Place is somewhere the app may have installed itself.
type Place struct {
	// Folder is the install folder, or the .app bundle on macOS.
	Folder string
	// Program is the program inside it.
	Program string
	// Bundle is true for a macOS .app, which is removed as a whole.
	Bundle bool
	// Shared is true when Folder also holds other programs (~/.local/bin), so only the
	// program is the app's own and the folder must never be removed.
	Shared bool
}

// Footprint is everything the installer can put on a machine, so that it can be taken
// away again and nothing else is.
type Footprint struct {
	Places []Place
	// Entries are the shortcuts and menu entries the installer writes (files, not folders).
	Entries []string
}

// footprintOf lists what installing on goos can leave behind. Unlike installTargets it
// includes Program Files, because an earlier run as an administrator may have used it.
func footprintOf(goos string, places installPlaces, startMenu, allUsersLink, applicationsDir string) Footprint {
	var footprint Footprint
	for _, target := range installTargets(goos, places) {
		footprint.Places = append(footprint.Places, Place{
			Folder: target.Folder, Program: target.Program,
			Bundle: target.Kind == kindMacBundle, Shared: target.Kind == kindLinuxUser,
		})
	}

	switch goos {
	case "windows":
		footprint.Entries = append(footprint.Entries, filepath.Join(startMenu, "MVD.lnk"), allUsersLink)
	case "linux":
		footprint.Entries = append(footprint.Entries, filepath.Join(applicationsDir, "mvd.desktop"), linuxIconPath(applicationsDir))
	}

	return footprint
}
