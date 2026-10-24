package folderdialog

import (
	"strings"
	"testing"
)

// A folder name is data. Whatever it contains, it must not end up in a script.
const hostile = `C:\x"; Remove-Item -Recurse C:\ #'$(calc)`

func TestMacPassesTheStartFolderAsAnArgumentAndOmitsItWhenThereIsNone(t *testing.T) {
	with, ok := folderDialogCommand("darwin", hostile, installed("osascript"))
	if !ok || with.Name != "osascript" {
		t.Fatalf("ok=%v cmd=%+v", ok, with)
	}
	if last := with.Args[len(with.Args)-1]; last != hostile {
		t.Errorf("the start folder is not the final argument: %q", last)
	}
	for _, a := range with.Args[:len(with.Args)-1] {
		if strings.Contains(a, "Remove-Item") {
			t.Errorf("the folder reached the script: %q", a)
		}
	}

	without, _ := folderDialogCommand("darwin", "", installed("osascript"))
	if len(without.Args) != len(with.Args)-1 {
		t.Errorf("an empty start should add no argument: %v", without.Args)
	}
}

func TestLinuxPrefersZenityThenKdialogThenNothing(t *testing.T) {
	both, ok := folderDialogCommand("linux", "/home/me/Music/", installed("kdialog", "zenity"))
	if !ok || both.Name != "zenity" {
		t.Fatalf("both installed: %+v", both)
	}
	if !contains(both.Args, "--filename=/home/me/Music/") {
		t.Errorf("zenity args = %v", both.Args)
	}

	kde, ok := folderDialogCommand("linux", "/home/me", installed("kdialog"))
	if !ok || kde.Name != "kdialog" || kde.Args[len(kde.Args)-1] != "/home/me" {
		t.Errorf("kdialog = %+v", kde)
	}

	if _, ok := folderDialogCommand("linux", "", installed()); ok {
		t.Error("expected no dialog when no tool is installed")
	}
}

func TestAMachineWithoutTheToolHasNoDialog(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		if _, ok := folderDialogCommand(goos, "", installed()); ok {
			t.Errorf("%s: found a dialog with nothing installed", goos)
		}
	}
}
