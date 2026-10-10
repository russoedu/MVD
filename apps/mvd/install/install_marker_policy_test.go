package install

import "testing"

func TestTheMarkerRemembersTheInstalledVersionAndTheDeclinedOne(t *testing.T) {
	state := markerState{InstalledVersion: "0.0.54", DeclinedUpdate: "0.0.62"}
	if got := parseMarker(state.String()); got != state {
		t.Errorf("got %+v, want %+v", got, state)
	}
}

func TestAMarkerFromAnOlderVersionOnlySaysAsked(t *testing.T) {
	if got := parseMarker("asked\n"); got != (markerState{}) {
		t.Errorf("got %+v", got)
	}
}

func TestVersionsAreComparedByNumber(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"0.0.54", "0.0.62", true},
		{"0.0.9", "0.0.10", true},
		{"0.0.62", "0.0.62", false},
		{"0.1.0", "0.0.62", false},
		{"dev", "0.0.62", false},
		{"", "0.0.62", false},
	} {
		if got := versionOlder(c.a, c.b); got != c.want {
			t.Errorf("versionOlder(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestAnUpdateIsOfferedOnlyWhenAnotherVersionIsInstalledAndTheRunningOneIsNotOlder(t *testing.T) {
	base := updateSituation{Version: "0.0.62", Window: true, HasInstalledCopy: true, Marker: markerState{InstalledVersion: "0.0.54"}}
	if !shouldOfferUpdate(base) {
		t.Error("a newer download than the installed copy should be offered")
	}

	legacy := base
	legacy.Marker = markerState{}
	if !shouldOfferUpdate(legacy) {
		t.Error("an installed copy of unknown version should be offered the update")
	}

	for name, change := range map[string]func(*updateSituation){
		"developer build":        func(s *updateSituation) { s.Version = "dev" },
		"no window":              func(s *updateSituation) { s.Window = false },
		"just moved":             func(s *updateSituation) { s.MovedFrom = `C:\old\mvd.exe` },
		"running the installed":  func(s *updateSituation) { s.Installed = true },
		"never moved":            func(s *updateSituation) { s.HasInstalledCopy = false },
		"declined this version":  func(s *updateSituation) { s.Marker.DeclinedUpdate = "0.0.62" },
		"same version installed": func(s *updateSituation) { s.Marker.InstalledVersion = "0.0.62" },
		"running an older one":   func(s *updateSituation) { s.Marker.InstalledVersion = "0.0.70" },
	} {
		s := base
		change(&s)
		if shouldOfferUpdate(s) {
			t.Errorf("%s: should not be offered", name)
		}
	}
}
