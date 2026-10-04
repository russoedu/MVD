package notification

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const hostileWords = `"; Remove-Item -Recurse C:\ #'$(calc) ` + "`n"

func TestWindowsShowsABalloonWithTheWordsInTheEnvironmentNotTheScript(t *testing.T) {
	notice := Notice{Title: hostileWords, Text: hostileWords}

	cmd, ok := notificationCommand("windows", notice, installed("powershell"))

	if !ok || cmd.Name != "powershell" {
		t.Fatalf("ok=%v cmd=%+v", ok, cmd)
	}
	if got := strings.Join(cmd.Args, " "); strings.Contains(got, "Remove-Item") || strings.Contains(got, "calc") {
		t.Errorf("the words reached the arguments: %s", got)
	}
	if decodePowerShell(t, cmd.Args) != windowsNoticeScript {
		t.Error("the encoded script is not the fixed script")
	}
	env := strings.Join(cmd.Env, "\n")
	if !strings.Contains(env, noticeTitleEnv+"=") || !strings.Contains(env, noticeTextEnv+"=") || !strings.Contains(env, noticeIconEnv+"=Info") {
		t.Errorf("env = %v", cmd.Env)
	}
}

func TestAFailureIsDrawnAsAnError(t *testing.T) {
	win, _ := notificationCommand("windows", Notice{Title: "t", Text: "x", Failure: true}, installed("powershell"))
	if !strings.Contains(strings.Join(win.Env, "\n"), noticeIconEnv+"=Error") {
		t.Errorf("windows env = %v", win.Env)
	}
	linux, _ := notificationCommand("linux", Notice{Title: "t", Text: "x", Failure: true}, installed("notify-send"))
	if !contains(linux.Args, "--urgency=critical") {
		t.Errorf("linux args = %v", linux.Args)
	}
}

func TestMacAndLinuxPassTheWordsAsArgumentsAfterEverythingThatIsCode(t *testing.T) {
	mac, ok := notificationCommand("darwin", Notice{Title: "Title", Text: hostileWords}, installed("osascript"))
	if !ok || mac.Name != "osascript" {
		t.Fatalf("mac = %+v", mac)
	}
	if n := len(mac.Args); mac.Args[n-1] != "Title" || mac.Args[n-2] != hostileWords {
		t.Errorf("mac args end with %q", mac.Args[len(mac.Args)-2:])
	}
	for _, a := range mac.Args[:len(mac.Args)-2] {
		if strings.Contains(a, "Remove-Item") {
			t.Errorf("the words reached the script: %q", a)
		}
	}

	linux, ok := notificationCommand("linux", Notice{Title: "Title", Text: hostileWords}, installed("notify-send"))
	if !ok || linux.Name != "notify-send" {
		t.Fatalf("linux = %+v", linux)
	}
	if n := len(linux.Args); linux.Args[n-3] != "--" || linux.Args[n-2] != "Title" || linux.Args[n-1] != hostileWords {
		t.Errorf("a title that starts with '-' could be read as an option; args = %v", linux.Args)
	}
}

func TestAMachineWithoutTheToolShowsNothing(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		if _, ok := notificationCommand(goos, Notice{Title: "t", Text: "x"}, installed()); ok {
			t.Errorf("%s: found a way to notify with nothing installed", goos)
		}
	}
}

func TestLongWordsAreCutToWhatABalloonCanShowWithoutBreakingACharacter(t *testing.T) {
	long := strings.Repeat("é", 400)

	cmd, _ := notificationCommand("windows", Notice{Title: long, Text: long}, installed("powershell"))

	for _, kv := range cmd.Env {
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
