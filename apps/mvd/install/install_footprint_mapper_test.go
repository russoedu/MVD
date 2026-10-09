package install

import (
	"path/filepath"
	"testing"
)

func TestWindowsCanHaveInstalledInEitherFolderAndLeftTwoShortcuts(t *testing.T) {
	places := installPlaces{ProgramFiles: filepath.Join("pf"), LocalAppData: filepath.Join("local")}

	got := footprintOf("windows", places, "menu", "all-users.lnk", "")

	if len(got.Places) != 2 || got.Places[0].Folder != filepath.Join("pf", "MVD") || got.Places[1].Folder != filepath.Join("local", "Programs", "MVD") {
		t.Errorf("places = %+v", got.Places)
	}
	if len(got.Entries) != 2 || got.Entries[0] != filepath.Join("menu", "MVD.lnk") || got.Entries[1] != "all-users.lnk" {
		t.Errorf("entries = %v", got.Entries)
	}
	for _, place := range got.Places {
		if place.Bundle || place.Shared {
			t.Errorf("a Windows install folder is the app's own: %+v", place)
		}
	}
}

func TestMacHasBundlesAndNoShortcuts(t *testing.T) {
	got := footprintOf("darwin", installPlaces{Home: "home"}, "", "", "")

	if len(got.Places) != 2 || !got.Places[0].Bundle || !got.Places[1].Bundle || got.Places[0].Folder != filepath.Join("/Applications", "MVD.app") {
		t.Errorf("places = %+v", got.Places)
	}
	if len(got.Entries) != 0 {
		t.Errorf("entries = %v", got.Entries)
	}
}

func TestOnLinuxTheBinFolderIsSharedAndTheMenuEntryIsTheOnlyExtra(t *testing.T) {
	got := footprintOf("linux", installPlaces{Home: "home"}, "", "", "apps")

	if len(got.Places) != 1 || !got.Places[0].Shared || got.Places[0].Program != filepath.Join("home", ".local", "bin", "mvd") {
		t.Errorf("places = %+v", got.Places)
	}
	if len(got.Entries) != 2 || got.Entries[0] != filepath.Join("apps", "mvd.desktop") || got.Entries[1] != linuxIconPath("apps") {
		t.Errorf("entries = %v", got.Entries)
	}
}
