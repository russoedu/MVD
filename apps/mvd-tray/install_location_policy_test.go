package main

import (
	"strings"
	"testing"
)

func TestOnWindowsTheAppBelongsInProgramsUnderLocalAppData(t *testing.T) {
	got := installFolder("windows", `C:\Users\me\AppData\Local`)

	if want := `C:\Users\me\AppData\Local\Programs\MVD`; !strings.EqualFold(strings.ReplaceAll(got, "/", `\`), want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestElsewhereTheAppHasNoFolderOfItsOwn(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		if got := installFolder(goos, "/home/me"); got != "" {
			t.Errorf("%s: got %q", goos, got)
		}
	}
	if got := installFolder("windows", "  "); got != "" {
		t.Errorf("without a local app data folder: got %q", got)
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
	return moveSituation{
		GOOS: "windows", Version: "0.0.7", Tray: true,
		ExeFolder: `C:\Users\me\Downloads`, Target: `C:\Users\me\AppData\Local\Programs\MVD`,
	}
}

func TestTheMoveIsOfferedWhenAReleaseRunsFromSomewhereElseForTheFirstTime(t *testing.T) {
	if !shouldOfferMove(situation()) {
		t.Error("expected the offer")
	}
}

func TestTheMoveIsNotOfferedInEachOfTheCasesWhereItWouldBeWrong(t *testing.T) {
	cases := map[string]func(*moveSituation){
		"there is no folder on this system": func(s *moveSituation) { s.Target = "" },
		"it is a developer's build":         func(s *moveSituation) { s.Version = "dev" },
		"it runs without a tray icon":       func(s *moveSituation) { s.Tray = false },
		"it is the copy a move just made":   func(s *moveSituation) { s.MovedFrom = `C:\Users\me\Downloads\mvd-tray.exe` },
		"the person was already asked":      func(s *moveSituation) { s.Asked = true },
		"it is already where it belongs":    func(s *moveSituation) { s.ExeFolder = `c:\users\me\appdata\local\programs\mvd` },
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

	if cmd.name != "powershell" {
		t.Fatalf("name = %q", cmd.name)
	}
	if got := strings.Join(cmd.args, " "); strings.Contains(got, "Remove-Item") {
		t.Errorf("the path reached the arguments: %s", got)
	}
	if decodePowerShell(t, cmd.args) != shortcutScript {
		t.Error("the encoded script is not the fixed script")
	}
	env := strings.Join(cmd.env, "\n")
	for _, want := range []string{`MVD_SHORTCUT=C:\Menu\MVD.lnk`, "MVD_TARGET=" + hostile} {
		if !strings.Contains(env, want) {
			t.Errorf("env is missing %q: %v", want, cmd.env)
		}
	}
}
