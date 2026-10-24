package folderdialog

import (
	"strings"
	"testing"
)

func TestMacAsksForTheFileTypeAndPassesTheStartFolderAsAnArgument(t *testing.T) {
	cmd, ok := fileDialogCommand("darwin", "/Users/me/plans", "mvd", "MVD review", installed("osascript"))
	if !ok || cmd.Name != "osascript" {
		t.Fatalf("got %+v, %v", cmd, ok)
	}
	if !contains(cmd.Args, "/Users/me/plans") {
		t.Errorf("start folder missing from %v", cmd.Args)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), `choose file of type {"mvd"}`) {
		t.Errorf("file type missing from %v", cmd.Args)
	}
}

func TestLinuxPrefersZenityThenKdialogThenNothingForFiles(t *testing.T) {
	cmd, ok := fileDialogCommand("linux", "/home/me", "mvd", "MVD review", installed("zenity", "kdialog"))
	if !ok || cmd.Name != "zenity" || !contains(cmd.Args, "--file-filter=MVD review | *.mvd") {
		t.Errorf("got %+v, %v", cmd, ok)
	}
	cmd, ok = fileDialogCommand("linux", "", "mvd", "MVD review", installed("kdialog"))
	if !ok || cmd.Name != "kdialog" || cmd.Args[0] != "--getopenfilename" {
		t.Errorf("got %+v, %v", cmd, ok)
	}
	if _, ok := fileDialogCommand("linux", "", "mvd", "MVD review", installed()); ok {
		t.Error("a machine without the tool has no window")
	}
}
