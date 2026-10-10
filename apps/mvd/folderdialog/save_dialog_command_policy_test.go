package folderdialog

import (
	"strings"
	"testing"
)

func TestWindowsOpensTheSaveWindowWithTheNameFolderAndFilterInTheEnvironment(t *testing.T) {
	cmd, ok := saveDialogCommand("windows", `C:\Users\me\plans`, "Dance Music.mvd", "mvd", "MVD review", installed("powershell"))
	if !ok || cmd.Name != "powershell" {
		t.Fatalf("got %+v, %v", cmd, ok)
	}
	for _, want := range []string{startEnv + `=C:\Users\me\plans`, nameEnv + "=Dance Music.mvd", filterEnv + "=MVD review (*.mvd)|*.mvd|All files (*.*)|*.*"} {
		if !contains(cmd.Env, want) {
			t.Errorf("%q missing from %v", want, cmd.Env)
		}
	}
	if script := decodePowerShell(t, cmd.Args); !strings.Contains(script, "SaveFileDialog") || !strings.Contains(script, "OverwritePrompt") {
		t.Errorf("not a save window that asks before replacing: %s", script)
	}
}

func TestMacPassesTheFolderAndTheNameAsArguments(t *testing.T) {
	cmd, ok := saveDialogCommand("darwin", "/Users/me/plans", "list.mvd", "mvd", "MVD review", installed("osascript"))
	if !ok || cmd.Name != "osascript" || !contains(cmd.Args, "/Users/me/plans") || !contains(cmd.Args, "list.mvd") {
		t.Errorf("got %+v, %v", cmd, ok)
	}
}

func TestLinuxSavesThroughZenityThenKdialogThenNothing(t *testing.T) {
	cmd, ok := saveDialogCommand("linux", "/home/me", "list.mvd", "mvd", "MVD review", installed("zenity", "kdialog"))
	if !ok || cmd.Name != "zenity" || !contains(cmd.Args, "--save") || !contains(cmd.Args, "--filename=/home/me/list.mvd") {
		t.Errorf("got %+v, %v", cmd, ok)
	}
	cmd, ok = saveDialogCommand("linux", "/home/me", "list.mvd", "mvd", "MVD review", installed("kdialog"))
	if !ok || cmd.Name != "kdialog" || cmd.Args[0] != "--getsavefilename" {
		t.Errorf("got %+v, %v", cmd, ok)
	}
	if _, ok := saveDialogCommand("linux", "", "", "mvd", "MVD review", installed()); ok {
		t.Error("a machine without the tool has no window")
	}
}

func TestTheExtensionIsAddedOnlyWhenTheNameHasNone(t *testing.T) {
	if got := withExtension("/a/b/list", "mvd"); got != "/a/b/list.mvd" {
		t.Errorf("got %q", got)
	}
	if got := withExtension("/a/b/list.mvd", "mvd"); got != "/a/b/list.mvd" {
		t.Errorf("got %q", got)
	}
	if got := withExtension("", "mvd"); got != "" {
		t.Errorf("got %q", got)
	}
}
