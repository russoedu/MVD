package main

import (
	"strings"
	"testing"
)

// A folder name is data. Whatever it contains, it must not end up in a script.
const hostile = `C:\x"; Remove-Item -Recurse C:\ #'$(calc)`

func TestWindowsUsesPowerShellWithTheStartFolderInTheEnvironmentNotTheScript(t *testing.T) {
	cmd, ok := folderDialogCommand("windows", hostile, installed("powershell"))

	if !ok || cmd.name != "powershell" {
		t.Fatalf("ok=%v cmd=%+v", ok, cmd)
	}
	if got := strings.Join(cmd.args, " "); strings.Contains(got, "Remove-Item") || strings.Contains(got, "calc") {
		t.Errorf("the folder reached the arguments: %s", got)
	}
	if len(cmd.env) != 1 || cmd.env[0] != "MVD_START="+hostile {
		t.Errorf("env = %v", cmd.env)
	}

	script := decodePowerShell(t, cmd.args)
	if script != windowsScript {
		t.Error("the encoded script is not the fixed script")
	}
	if !strings.Contains(strings.Join(cmd.args, " "), "-STA") {
		t.Error("a Windows dialog needs a single-threaded apartment")
	}
}

func TestMacPassesTheStartFolderAsAnArgumentAndOmitsItWhenThereIsNone(t *testing.T) {
	with, ok := folderDialogCommand("darwin", hostile, installed("osascript"))
	if !ok || with.name != "osascript" {
		t.Fatalf("ok=%v cmd=%+v", ok, with)
	}
	if last := with.args[len(with.args)-1]; last != hostile {
		t.Errorf("the start folder is not the final argument: %q", last)
	}
	for _, a := range with.args[:len(with.args)-1] {
		if strings.Contains(a, "Remove-Item") {
			t.Errorf("the folder reached the script: %q", a)
		}
	}

	without, _ := folderDialogCommand("darwin", "", installed("osascript"))
	if len(without.args) != len(with.args)-1 {
		t.Errorf("an empty start should add no argument: %v", without.args)
	}
}

func TestLinuxPrefersZenityThenKdialogThenNothing(t *testing.T) {
	both, ok := folderDialogCommand("linux", "/home/me/Music/", installed("kdialog", "zenity"))
	if !ok || both.name != "zenity" {
		t.Fatalf("both installed: %+v", both)
	}
	if !contains(both.args, "--filename=/home/me/Music/") {
		t.Errorf("zenity args = %v", both.args)
	}

	kde, ok := folderDialogCommand("linux", "/home/me", installed("kdialog"))
	if !ok || kde.name != "kdialog" || kde.args[len(kde.args)-1] != "/home/me" {
		t.Errorf("kdialog = %+v", kde)
	}

	if _, ok := folderDialogCommand("linux", "", installed()); ok {
		t.Error("expected no dialog when no tool is installed")
	}
}

func TestAMachineWithoutTheToolHasNoDialog(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		if _, ok := folderDialogCommand(goos, "", installed()); ok {
			t.Errorf("%s: found a dialog with nothing installed", goos)
		}
	}
}
