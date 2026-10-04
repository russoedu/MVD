package macbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestTheBundleHoldsTheProgramAndItsInfoPlist(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "MVD.app")

	if err := Write(bundle, writeFile(t, "program", "binary"), "1.2.3", ""); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(ProgramPath(bundle)); string(data) != "binary" {
		t.Errorf("program = %q", data)
	}
	plist, _ := os.ReadFile(InfoPlistPath(bundle))
	if !strings.Contains(string(plist), "<string>1.2.3</string>") || strings.Contains(string(plist), "CFBundleIconFile") {
		t.Errorf("Info.plist = %s", plist)
	}
	if _, err := os.Stat(IconPath(bundle)); err == nil {
		t.Error("an icon appeared although none was given")
	}
}

func TestAnIconGivenIsPlacedInResourcesAndNamedInTheInfoPlist(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "MVD.app")

	if err := Write(bundle, writeFile(t, "program", "binary"), "1", writeFile(t, "logo.icns", "icon")); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(IconPath(bundle)); string(data) != "icon" {
		t.Errorf("icon = %q", data)
	}
	if plist, _ := os.ReadFile(InfoPlistPath(bundle)); !strings.Contains(string(plist), "CFBundleIconFile") {
		t.Error("the Info.plist does not name the icon")
	}
}

func TestWritingOverAnOlderBundleReplacesIt(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "MVD.app")
	if err := Write(bundle, writeFile(t, "old", "old"), "0.0.1", ""); err != nil {
		t.Fatal(err)
	}

	if err := Write(bundle, writeFile(t, "new", "new"), "0.0.2", ""); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(ProgramPath(bundle)); string(data) != "new" {
		t.Errorf("program = %q", data)
	}
	if plist, _ := os.ReadFile(InfoPlistPath(bundle)); !strings.Contains(string(plist), "0.0.2") {
		t.Error("the Info.plist was not replaced")
	}
}

func TestMissingPiecesAreReportedNotIgnored(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "MVD.app")
	dir := t.TempDir()

	if err := Write(bundle, filepath.Join(dir, "missing"), "1", ""); err == nil {
		t.Error("a missing program was not reported")
	}
	if err := Write(bundle, writeFile(t, "program", "binary"), "1", filepath.Join(dir, "missing.icns")); err == nil {
		t.Error("a missing icon was not reported")
	}
}

func TestPathsInsideTheBundleFollowTheMacOSLayout(t *testing.T) {
	root := filepath.Join("a", "MVD.app")

	if got, want := ProgramPath(root), filepath.Join(root, "Contents", "MacOS", "mvd-tray"); got != want {
		t.Errorf("program path = %s, want %s", got, want)
	}
	if got, want := InfoPlistPath(root), filepath.Join(root, "Contents", "Info.plist"); got != want {
		t.Errorf("Info.plist path = %s, want %s", got, want)
	}
	if got, want := IconPath(root), filepath.Join(root, "Contents", "Resources", "MVD.icns"); got != want {
		t.Errorf("icon path = %s, want %s", got, want)
	}
}
