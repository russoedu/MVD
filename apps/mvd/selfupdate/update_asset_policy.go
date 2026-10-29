package selfupdate

import "fmt"

// programName is the file of the program inside a release's zip.
const programName = "mvd.exe"

// assetName is the file of a release that holds the program for this system, and ok is false
// where MVD does not replace itself: only Windows does so far. The other systems are sent to
// the release page, where their builds are.
func assetName(version, goos, goarch string) (name string, ok bool) {
	if goos != "windows" || goarch != "amd64" {
		return "", false
	}

	return fmt.Sprintf("mvd_%s_windows_amd64.zip", version), true
}

// assetOf finds the asset called name in a release.
func assetOf(r Release, name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}

	return Asset{}, false
}
