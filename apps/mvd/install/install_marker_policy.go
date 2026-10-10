package install

import (
	"strconv"
	"strings"
)

// markerState is what the "install-offered" file in the app-data folder remembers: that
// the person has been asked to move the app, which version was put in place, and which
// version they said no to updating it with. A file written by an older version only says
// "asked", which leaves both versions empty.
type markerState struct {
	// InstalledVersion is the version of the program the app put in its own folder; ""
	// when it is not known.
	InstalledVersion string
	// DeclinedUpdate is the running version the person said "Not now" to updating the
	// installed copy with, so that it is not asked about again for that same version.
	DeclinedUpdate string
}

// parseMarker reads the file's text.
func parseMarker(text string) markerState {
	var state markerState
	for _, line := range strings.Split(text, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found {
			continue
		}
		switch key {
		case "installed":
			state.InstalledVersion = strings.TrimSpace(value)
		case "declined":
			state.DeclinedUpdate = strings.TrimSpace(value)
		}
	}

	return state
}

// String is the file's text.
func (s markerState) String() string {
	text := "asked\n"
	if s.InstalledVersion != "" {
		text += "installed " + s.InstalledVersion + "\n"
	}
	if s.DeclinedUpdate != "" {
		text += "declined " + s.DeclinedUpdate + "\n"
	}

	return text
}

// versionOlder reports whether version a is older than b, for versions written
// "major.minor.patch". A version that cannot be read is never called older, so the
// person is asked rather than left on a copy that is out of date.
func versionOlder(a, b string) bool {
	pa, okA := versionParts(a)
	pb, okB := versionParts(b)
	if !okA || !okB {
		return false
	}
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}

	return false
}

func versionParts(version string) (parts [3]int, ok bool) {
	pieces := strings.Split(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".")
	if len(pieces) != 3 {
		return parts, false
	}
	for i, piece := range pieces {
		n, err := strconv.Atoi(piece)
		if err != nil {
			return parts, false
		}
		parts[i] = n
	}

	return parts, true
}

// updateSituation is everything that decides whether to offer updating the copy of the
// app that is already in its own folder with the one that is running.
type updateSituation struct {
	// Version is the running version, "dev" for a build made on a developer's machine.
	Version string
	// Window is false when the app was asked to run without a window.
	Window bool
	// MovedFrom is set on the copy that a move has just started.
	MovedFrom string
	// Installed is true when the running program is itself the one in its folder.
	Installed bool
	// HasInstalledCopy is true when a program exists in one of the app's own folders.
	HasInstalledCopy bool
	// Marker is what the file remembers.
	Marker markerState
}

// shouldOfferUpdate reports whether to ask to replace the installed copy with the running
// program: when a copy is installed, this is not it, and it is not the same version, nor an
// older one, nor one the person already said "Not now" to. Someone who never moved the app
// has no installed copy, so is not asked.
func shouldOfferUpdate(s updateSituation) bool {
	switch {
	case s.Version == "dev", !s.Window, s.MovedFrom != "", s.Installed, !s.HasInstalledCopy:
		return false
	case s.Marker.DeclinedUpdate == s.Version:
		return false
	case s.Marker.InstalledVersion == s.Version:
		return false
	case s.Marker.InstalledVersion != "" && versionOlder(s.Version, s.Marker.InstalledVersion):
		return false
	}

	return true
}
