package macbundle

import "path/filepath"

// ExecutableName is the program's file name inside the bundle, and what the Info.plist
// tells macOS to start.
const ExecutableName = "mvd"

// IconName is the icon file's name without its extension, as the Info.plist gives it.
const IconName = "MVD"

// ProgramPath is where the program sits inside the bundle at bundle.
func ProgramPath(bundle string) string {
	return filepath.Join(bundle, "Contents", "MacOS", ExecutableName)
}

// InfoPlistPath is where the Info.plist sits inside the bundle at bundle.
func InfoPlistPath(bundle string) string {
	return filepath.Join(bundle, "Contents", "Info.plist")
}

// IconPath is where the icon sits inside the bundle at bundle.
func IconPath(bundle string) string {
	return filepath.Join(bundle, "Contents", "Resources", IconName+".icns")
}
