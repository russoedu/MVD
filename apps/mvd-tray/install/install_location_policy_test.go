package install

import (
	"path/filepath"
	"strings"
	"testing"
)

var places = installPlaces{
	ProgramFiles: filepath.Join("C:", "Program Files"),
	LocalAppData: filepath.Join("C:", "Users", "me", "AppData", "Local"),
	Home:         filepath.Join("home", "me"),
}

func TestWindowsOffersProgramFilesForEveryoneThenTheOwnProgramsFolder(t *testing.T) {
	targets := installTargets("windows", places)

	if len(targets) != 2 {
		t.Fatalf("targets = %+v", targets)
	}
	system, user := targets[0], targets[1]
	if system.Kind != kindWindowsSystem || !system.Everyone || system.Folder != filepath.Join(places.ProgramFiles, "MVD") {
		t.Errorf("system = %+v", system)
	}
	if system.Program != filepath.Join(system.Folder, "mvd-tray.exe") {
		t.Errorf("system program = %s", system.Program)
	}
	if user.Kind != kindWindowsUser || user.Everyone || user.Folder != filepath.Join(places.LocalAppData, "Programs", "MVD") {
		t.Errorf("user = %+v", user)
	}
}

func TestWindowsWithoutProgramFilesStillHasTheOwnFolder(t *testing.T) {
	targets := installTargets("windows", installPlaces{LocalAppData: places.LocalAppData})

	if len(targets) != 1 || targets[0].Kind != kindWindowsUser {
		t.Errorf("targets = %+v", targets)
	}
}

func TestMacOffersApplicationsForEveryoneThenTheOwnApplicationsFolderAsBundles(t *testing.T) {
	targets := installTargets("darwin", places)

	if len(targets) != 2 {
		t.Fatalf("targets = %+v", targets)
	}
	system, user := targets[0], targets[1]
	if system.Kind != kindMacBundle || !system.Everyone || system.Folder != filepath.Join("/Applications", "MVD.app") {
		t.Errorf("system = %+v", system)
	}
	if want := filepath.Join(system.Folder, "Contents", "MacOS", "mvd-tray"); system.Program != want {
		t.Errorf("system program = %s, want %s", system.Program, want)
	}
	if user.Everyone || user.Folder != filepath.Join(places.Home, "Applications", "MVD.app") {
		t.Errorf("user = %+v", user)
	}
}

func TestLinuxHasOnlyTheOwnLocalBinFolder(t *testing.T) {
	targets := installTargets("linux", places)

	if len(targets) != 1 || targets[0].Kind != kindLinuxUser || targets[0].Everyone {
		t.Fatalf("targets = %+v", targets)
	}
	if want := filepath.Join(places.Home, ".local", "bin", "mvd-tray"); targets[0].Program != want {
		t.Errorf("program = %s, want %s", targets[0].Program, want)
	}
}

func TestAMachineWithNowhereToPutItGetsNoTargets(t *testing.T) {
	if got := installTargets("linux", installPlaces{}); len(got) != 0 {
		t.Errorf("linux without a home: %+v", got)
	}
	if got := installTargets("freebsd", places); len(got) != 0 {
		t.Errorf("an unknown system: %+v", got)
	}
}

func TestAProgramInAnyOfTheTargetFoldersCountsAsInstalledEvenIfRenamed(t *testing.T) {
	targets := installTargets("windows", places)

	for _, folder := range []string{targets[0].Folder, targets[1].Folder} {
		if !isInstalled("windows", filepath.Join(folder, "mvd-tray (1).exe"), targets) {
			t.Errorf("%s was not recognised as installed", folder)
		}
	}
	if isInstalled("windows", filepath.Join("C:", "Users", "me", "Downloads", "mvd-tray.exe"), targets) {
		t.Error("Downloads is not an install folder")
	}
}

func TestInsideAnAppBundleCountsAsInstalledOnMac(t *testing.T) {
	targets := installTargets("darwin", places)

	if !isInstalled("darwin", targets[1].Program, targets) || !isInstalled("darwin", targets[0].Program, targets) {
		t.Error("the programs inside the bundles were not recognised as installed")
	}
}

func TestWindowsPathsAreComparedWithoutRegardToCase(t *testing.T) {
	// Forward slashes, which every host understands: a backslash is only a separator on Windows.
	if !samePlace("windows", "C:/Users/Me/AppData/Local/Programs/MVD", "c:/users/me/appdata/local/programs/mvd/") {
		t.Error("the same Windows folder in another case and with a trailing slash should match")
	}
	if samePlace("windows", "C:/Users/me/Downloads", "C:/Users/me/AppData/Local/Programs/MVD") {
		t.Error("different folders should not match")
	}
	if samePlace("linux", "/opt/MVD", "/opt/mvd") {
		t.Error("Linux paths are case sensitive")
	}
}

func situation() moveSituation {
	return moveSituation{Version: "0.0.7", Tray: true, HasTarget: true}
}

func TestTheMoveIsOfferedWhenAReleaseRunsFromSomewhereElseForTheFirstTime(t *testing.T) {
	if !shouldOfferMove(situation()) {
		t.Error("expected the offer")
	}
}

func TestTheMoveIsNotOfferedInEachOfTheCasesWhereItWouldBeWrong(t *testing.T) {
	cases := map[string]func(*moveSituation){
		"there is nowhere to move to":     func(s *moveSituation) { s.HasTarget = false },
		"it is a developer's build":       func(s *moveSituation) { s.Version = "dev" },
		"it runs without a tray icon":     func(s *moveSituation) { s.Tray = false },
		"it is the copy a move just made": func(s *moveSituation) { s.MovedFrom = "mvd-tray.exe" },
		"the person was already asked":    func(s *moveSituation) { s.Asked = true },
		"it is already installed":         func(s *moveSituation) { s.Installed = true },
	}
	for name, change := range cases {
		s := situation()
		change(&s)

		if shouldOfferMove(s) {
			t.Errorf("offered although %s", name)
		}
	}
}

func TestTheShortcutCommandPassesPathsInTheEnvironmentNotTheScript(t *testing.T) {
	hostile := `C:\x"; Remove-Item -Recurse C:\ #'$(calc)\mvd-tray.exe`

	cmd := shortcutCommand(`C:\Menu\MVD.lnk`, hostile)

	if cmd.Name != "powershell" {
		t.Fatalf("name = %q", cmd.Name)
	}
	if got := strings.Join(cmd.Args, " "); strings.Contains(got, "Remove-Item") {
		t.Errorf("the path reached the arguments: %s", got)
	}
	if decodePowerShell(t, cmd.Args) != shortcutScript {
		t.Error("the encoded script is not the fixed script")
	}
	env := strings.Join(cmd.Env, "\n")
	for _, want := range []string{`MVD_SHORTCUT=C:\Menu\MVD.lnk`, "MVD_TARGET=" + hostile} {
		if !strings.Contains(env, want) {
			t.Errorf("env is missing %q: %v", want, cmd.Env)
		}
	}
}
