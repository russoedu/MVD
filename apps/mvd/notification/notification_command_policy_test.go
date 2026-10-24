package notification

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const hostileWords = `"; Remove-Item -Recurse C:\ #'$(calc) ` + "`n"

func TestWindowsStartsNoProgramToShowANotice(t *testing.T) {
	if cmd, ok := notificationCommand("windows", Notice{Title: hostileWords, Text: hostileWords}, installed("powershell")); ok {
		t.Errorf("a program is started: %+v", cmd)
	}
}

func TestAFailureIsUrgentOnLinux(t *testing.T) {
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
	for _, goos := range []string{"darwin", "linux"} {
		if _, ok := notificationCommand(goos, Notice{Title: "t", Text: "x"}, installed()); ok {
			t.Errorf("%s: found a way to notify with nothing installed", goos)
		}
	}
}

func TestLongWordsAreCutToWhatABalloonCanShowWithoutBreakingACharacter(t *testing.T) {
	long := strings.Repeat("é", 400)

	for text, limit := range map[string]int{clip(long, noticeTitleLimit): noticeTitleLimit, clip(long, noticeTextLimit): noticeTextLimit} {
		if got := utf8.RuneCountInString(text); got != limit {
			t.Errorf("%d characters, want %d", got, limit)
		}
		if !strings.HasSuffix(text, "...") || !utf8.ValidString(text) {
			t.Errorf("not cut cleanly: %q", text)
		}
	}
}
