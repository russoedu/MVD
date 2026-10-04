package folderdialog

import "testing"

func TestFolderChoiceTrimsTheOutputAndTheTrailingSeparator(t *testing.T) {
	cases := map[string]string{
		"/home/me/Music\n":       "/home/me/Music",
		"/Users/me/Music/\n":     "/Users/me/Music",
		"  /a/b//  ":             "/a/b",
		"C:\\Users\\me\\Music":   "C:\\Users\\me\\Music",
		"C:\\Users\\me\\Music\\": "C:\\Users\\me\\Music",
		"/":                      "/",
		"C:\\":                   "C:\\",
		"/música/日本\n":           "/música/日本",
	}
	for output, want := range cases {
		got, chosen := folderChoice(output)
		if !chosen || got != want {
			t.Errorf("%q -> %q (chosen=%v), want %q", output, got, chosen, want)
		}
	}
}

func TestEmptyOutputIsACancel(t *testing.T) {
	for _, output := range []string{"", "\n", "   \r\n"} {
		if path, chosen := folderChoice(output); chosen || path != "" {
			t.Errorf("%q -> %q (chosen=%v)", output, path, chosen)
		}
	}
}

func TestOnlyAnExitOfOneWithNothingToSayIsACancelOnLinux(t *testing.T) {
	if !cancelledByExit("linux", 1, "") {
		t.Error("exit 1 with no message is zenity/kdialog being closed")
	}
	if cancelledByExit("linux", 1, "Gtk-WARNING: cannot open display") {
		t.Error("a message means it failed, not that it was cancelled")
	}
	if cancelledByExit("linux", 5, "") || cancelledByExit("linux", 0, "") {
		t.Error("only exit 1 is a cancel")
	}
}

func TestMacRecognisesTheCancelError(t *testing.T) {
	if !cancelledByExit("darwin", 1, "execution error: User canceled. (-128)") {
		t.Error("user cancelled was not recognised")
	}
	if cancelledByExit("darwin", 1, "execution error: something else (-1700)") {
		t.Error("another error was taken for a cancel")
	}
}

func TestOnWindowsACancelIsEmptyOutputNeverAnExitCode(t *testing.T) {
	if cancelledByExit("windows", 1, "") {
		t.Error("a PowerShell failure must not look like a cancel")
	}
}
