package question

import (
	"strings"
	"testing"
)

const hostileQuestion = `"; do shell script "rm -rf ~" --`

func TestMacAsksThroughOsascriptWithEverythingTheWordsAreAsArguments(t *testing.T) {
	choices := []string{"For everyone", "Just for me", "Leave it here"}

	cmd, ok := questionCommand("darwin", "MVD", hostileQuestion, choices, installed("osascript"))

	if !ok || cmd.Name != "osascript" {
		t.Fatalf("ok=%v cmd=%+v", ok, cmd)
	}
	script := strings.Join(cmd.Args[:8], " ")
	if strings.Contains(script, "rm -rf") {
		t.Errorf("the text reached the script: %s", script)
	}
	// After the script: the text, the title, then the labels in reverse.
	tail := cmd.Args[8:]
	want := []string{hostileQuestion, "MVD", "Leave it here", "Just for me", "For everyone"}
	if strings.Join(tail, "|") != strings.Join(want, "|") {
		t.Errorf("arguments after the script = %q, want %q", tail, want)
	}
	if !strings.Contains(script, "default button (item 5 of argv)") {
		t.Errorf("the preferred choice (the last label, because they are reversed) should be the default: %s", script)
	}
}

func TestMacWithTwoChoicesUsesTwoButtonsAndTheRightDefault(t *testing.T) {
	cmd, _ := questionCommand("darwin", "MVD", "Move it?", []string{"Move it", "Leave it here"}, installed("osascript"))

	script := strings.Join(cmd.Args[:8], " ")
	if !strings.Contains(script, "buttons {item 3 of argv, item 4 of argv}") || !strings.Contains(script, "default button (item 4 of argv)") {
		t.Errorf("script = %s", script)
	}
}

func TestLinuxAsksWithZenityElseKdialogAndTheLabelsAreArguments(t *testing.T) {
	choices := []string{"Move it", "Leave it here"}

	zenity, ok := questionCommand("linux", "MVD", "Move?", choices, installed("kdialog", "zenity"))
	if !ok || zenity.Name != "zenity" || !contains(zenity.Args, "--ok-label=Move it") || !contains(zenity.Args, "--cancel-label=Leave it here") {
		t.Errorf("zenity = %+v", zenity)
	}

	kde, ok := questionCommand("linux", "MVD", "Move?", choices, installed("kdialog"))
	if !ok || kde.Name != "kdialog" || kde.Args[len(kde.Args)-2] != "--yesno" {
		t.Errorf("kdialog = %+v", kde)
	}
}

func TestThereIsNoQuestionWhenNothingCanAskIt(t *testing.T) {
	two := []string{"Move it", "Leave it here"}
	if _, ok := questionCommand("linux", "MVD", "x", two, installed()); ok {
		t.Error("Linux with no dialog program")
	}
	if _, ok := questionCommand("darwin", "MVD", "x", two, installed()); ok {
		t.Error("macOS without osascript")
	}
	if _, ok := questionCommand("linux", "MVD", "x", []string{"a", "b", "c"}, installed("zenity")); ok {
		t.Error("Linux has no three-way box and never needs one")
	}
	if _, ok := questionCommand("darwin", "MVD", "x", []string{"only"}, installed("osascript")); ok {
		t.Error("a question needs at least two choices")
	}
}

func TestTheMacAnswerIsTheButtonThatWasPressed(t *testing.T) {
	choices := []string{"For everyone", "Just for me", "Leave it here"}
	cases := []struct {
		stdout string
		code   int
		want   Answer
	}{
		{"For everyone\n", 0, AnswerFirst},
		{"Just for me\n", 0, AnswerSecond},
		{"Leave it here\n", 0, AnswerLeave},
		{"", 1, AnswerLeave},
		{"something else\n", 0, AnswerLeave},
	}
	for _, c := range cases {
		if got := parseAnswer("darwin", choices, c.code, c.stdout); got != c.want {
			t.Errorf("%q (exit %d): got %v, want %v", c.stdout, c.code, got, c.want)
		}
	}
}

func TestTheLinuxAnswerIsTheExitStatus(t *testing.T) {
	choices := []string{"Move it", "Leave it here"}
	for code, want := range map[int]Answer{0: AnswerFirst, 1: AnswerLeave, 5: AnswerUnavailable, 255: AnswerUnavailable} {
		if got := parseAnswer("linux", choices, code, ""); got != want {
			t.Errorf("exit %d: got %v, want %v", code, got, want)
		}
	}
}
