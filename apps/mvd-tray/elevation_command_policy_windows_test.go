package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This runs the very script that is run with administrator rights, but without them,
// against folders that need none, so what it does is checked for real: that it copies the
// program, makes the shortcut, and that a quote in a path does not break out of its string.
func TestTheElevatedScriptReallyInstallsAndMakesTheShortcutEvenFromAFolderWithAQuoteInItsName(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "it's downloaded", "mvd-tray (1).exe")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("the program"), 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "Program Files", "MVD", "mvd-tray.exe")
	link := filepath.Join(root, "ProgramData", "Start Menu", "Programs", "MVD.lnk")

	script := encodePowerShell(elevatedInstallScript(source, dest, link))
	output, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-EncodedCommand", script).CombinedOutput()

	if err != nil {
		t.Fatalf("the script failed: %v\n%s", err, output)
	}
	if got, _ := os.ReadFile(dest); string(got) != "the program" {
		t.Errorf("installed program = %q", got)
	}
	if _, err := os.Stat(link); err != nil {
		t.Errorf("the shortcut was not made: %v", err)
	}
}

func TestTheElevatedScriptFailsWithANonZeroExitWhenThereIsNothingToCopy(t *testing.T) {
	root := t.TempDir()
	script := encodePowerShell(elevatedInstallScript(filepath.Join(root, "missing.exe"), filepath.Join(root, "out", "mvd-tray.exe"), filepath.Join(root, "l.lnk")))

	err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-EncodedCommand", script).Run()

	if err == nil {
		t.Error("expected a non-zero exit so the app knows the install failed")
	}
}
