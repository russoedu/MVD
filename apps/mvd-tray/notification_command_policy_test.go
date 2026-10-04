package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const hostileWords = `"; Remove-Item -Recurse C:\ #'$(calc) ` + "`n"

func TestWindowsShowsABalloonWithTheWordsInTheEnvironmentNotTheScript(t *testing.T) {
	notice := userNotice{Title: hostileWords, Text: hostileWords}

	cmd, ok := notificationCommand("windows", notice, installed("powershell"))

	if !ok || cmd.name != "powershell" {
		t.Fatalf("ok=%v cmd=%+v", ok, cmd)
	}
	if got := strings.Join(cmd.args, " "); strings.Contains(got, "Remove-Item") || strings.Contains(got, "calc") {
		t.Errorf("the words reached the arguments: %s", got)
	}
	if decodePowerShell(t, cmd.args) != windowsNoticeScript {
		t.Error("the encoded script is not the fixed script")
	}
	env := strings.Join(cmd.env, "\n")
	if !strings.Contains(env, noticeTitleEnv+"=") || !strings.Contains(env, noticeTextEnv+"=") || !strings.Contains(env, noticeIconEnv+"=Info") {
		t.Errorf("env = %v", cmd.env)
	}
}

func TestAFailureIsDrawnAsAnError(t *testing.T) {
	win, _ := notificationCommand("windows", userNotice{Title: "t", Text: "x", Failure: true}, installed("powershell"))
	if !strings.Contains(strings.Join(win.env, "\n"), noticeIconEnv+"=Error") {
		t.Errorf("windows env = %v", win.env)
	}
	linux, _ := notificationCommand("linux", userNotice{Title: "t", Text: "x", Failure: true}, installed("notify-send"))
	if !contains(linux.args, "--urgency=critical") {
		t.Errorf("linux args = %v", linux.args)
	}
}

func TestMacAndLinuxPassTheWordsAsArgumentsAfterEverythingThatIsCode(t *testing.T) {
	mac, ok := notificationCommand("darwin", userNotice{Title: "Title", Text: hostileWords}, installed("osascript"))
	if !ok || mac.name != "osascript" {
		t.Fatalf("mac = %+v", mac)
	}
	if n := len(mac.args); mac.args[n-1] != "Title" || mac.args[n-2] != hostileWords {
		t.Errorf("mac args end with %q", mac.args[len(mac.args)-2:])
	}
	for _, a := range mac.args[:len(mac.args)-2] {
		if strings.Contains(a, "Remove-Item") {
			t.Errorf("the words reached the script: %q", a)
		}
	}

	linux, ok := notificationCommand("linux", userNotice{Title: "Title", Text: hostileWords}, installed("notify-send"))
	if !ok || linux.name != "notify-send" {
		t.Fatalf("linux = %+v", linux)
	}
	if n := len(linux.args); linux.args[n-3] != "--" || linux.args[n-2] != "Title" || linux.args[n-1] != hostileWords {
		t.Errorf("a title that starts with '-' could be read as an option; args = %v", linux.args)
	}
}

func TestAMachineWithoutTheToolShowsNothing(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		if _, ok := notificationCommand(goos, userNotice{Title: "t", Text: "x"}, installed()); ok {
			t.Errorf("%s: found a way to notify with nothing installed", goos)
		}
	}
}

func TestLongWordsAreCutToWhatABalloonCanShowWithoutBreakingACharacter(t *testing.T) {
	long := strings.Repeat("é", 400)

	cmd, _ := notificationCommand("windows", userNotice{Title: long, Text: long}, installed("powershell"))

	for _, kv := range cmd.env {
		name, value, _ := strings.Cut(kv, "=")
		limit := map[string]int{noticeTitleEnv: noticeTitleLimit, noticeTextEnv: noticeTextLimit}[name]
		if limit == 0 {
			continue
		}
		if got := utf8.RuneCountInString(value); got != limit {
			t.Errorf("%s has %d characters, want %d", name, got, limit)
		}
		if !strings.HasSuffix(value, "...") || !utf8.ValidString(value) {
			t.Errorf("%s was not cut cleanly: %q", name, value)
		}
	}
}
