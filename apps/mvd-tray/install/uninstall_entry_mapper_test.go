package install

import (
	"path/filepath"
	"testing"
)

func TestTheEntryPointsWindowsAtTheInstalledProgramWithTheUninstallFlag(t *testing.T) {
	target := installTarget{Folder: filepath.Join("C:", "Users", "a", "MVD"), Program: filepath.Join("C:", "Users", "a", "MVD", "mvd-tray.exe")}

	entry := uninstallEntryFor(target, "1.2.3")

	if entry.UninstallString != `"`+target.Program+`" -uninstall` {
		t.Errorf("UninstallString = %s", entry.UninstallString)
	}
	want := map[string]string{
		"DisplayName": "MVD", "DisplayVersion": "1.2.3", "Publisher": "MVD",
		"InstallLocation": target.Folder, "DisplayIcon": target.Program,
	}
	for name, value := range want {
		if got := entry.strings()[name]; got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
}

func TestTheEntryOffersRemovalOnlyNotModifyingOrRepairing(t *testing.T) {
	numbers := uninstallEntryFor(installTarget{}, "1").numbers()

	if numbers["NoModify"] != 1 || numbers["NoRepair"] != 1 {
		t.Errorf("numbers = %v", numbers)
	}
}
