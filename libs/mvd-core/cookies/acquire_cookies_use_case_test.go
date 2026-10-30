package cookies

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain lets the test binary double as a yt-dlp stub for cookie export.
func TestMain(m *testing.M) {
	if os.Getenv("MVD_STUB_YTDLP") == "1" {
		os.Exit(stubExport(os.Args[1:]))
	}
	os.Exit(m.Run())
}

// stubExport mimics `yt-dlp --cookies-from-browser B --cookies FILE ...`:
// edge fails to decrypt, firefox exports only a consent cookie, chrome
// exports a real session, opera a real session of a Premium account. A "-F" call lists the
// formats of the cookie file it is given: the Premium audio only for a Premium login.
func stubExport(args []string) int {
	browser, file, list := "", "", false
	for i, a := range args {
		switch a {
		case "-F":
			list = true
		case "--cookies-from-browser":
			browser = args[i+1]
		case "--cookies":
			file = args[i+1]
		}
	}
	if list {
		data, _ := os.ReadFile(file)
		fmt.Println("140 m4a audio only 130k")
		if strings.Contains(string(data), "PREMIUM") {
			fmt.Println("141 m4a audio only 258k")
		}
		return 0
	}
	switch browser {
	case "opera":
		_ = os.WriteFile(file, []byte("# Netscape HTTP Cookie File\n.google.com\tTRUE\t/\tTRUE\t0\tSID\tsecret\n.youtube.com\tTRUE\t/\tTRUE\t0\tPREMIUM\tyes\n"), 0600)
		return 0
	case "edge":
		fmt.Fprintln(os.Stderr, "ERROR: could not copy edge cookie database")
		return 1
	case "chrome":
		_ = os.WriteFile(file, []byte("# Netscape HTTP Cookie File\n.google.com\tTRUE\t/\tTRUE\t0\tSID\tsecret\n.youtube.com\tTRUE\t/\tTRUE\t0\tLOGIN_INFO\tyes\n"), 0600)
		return 0
	default: // firefox and others: consent cookie only, no login
		_ = os.WriteFile(file, []byte("# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tFALSE\t0\tPREF\thl=en\n"), 0600)
		return 0
	}
}

func useStub(t *testing.T) string {
	t.Helper()
	t.Setenv("MVD_STUB_YTDLP", "1")
	return os.Args[0]
}

func TestAcquire(t *testing.T) {
	bin := useStub(t)
	file := filepath.Join(t.TempDir(), "cookies.txt")

	// firefox (no login) and edge (unreadable) are skipped; chrome wins.
	b, ok := Acquire(context.Background(), bin, file, "https://youtube.com/playlist?list=A", nil, []string{"firefox", "edge", "chrome"}, nil)
	if !ok || b != "chrome" {
		t.Fatalf("want chrome, got %q ok=%v", b, ok)
	}
	if data, _ := os.ReadFile(file); !strings.Contains(string(data), "SID") {
		t.Errorf("chrome cookies not kept: %q", data)
	}

	// Only a browser without a login: not ok, and the junk file is removed.
	file2 := filepath.Join(t.TempDir(), "cookies.txt")
	if _, ok := Acquire(context.Background(), bin, file2, "p", nil, []string{"firefox"}, nil); ok {
		t.Error("firefox has no session; should be not ok")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("a failed acquire should leave no cookie file")
	}

	// No browsers at all.
	if _, ok := Acquire(context.Background(), bin, file2, "p", nil, nil, nil); ok {
		t.Error("no browsers should be not ok")
	}
}

func TestAcquirePrefersThePremiumLogin(t *testing.T) {
	bin := useStub(t)
	file := filepath.Join(t.TempDir(), "cookies.txt")

	// chrome is logged in without Premium and comes first; opera is Premium and wins.
	b, ok := Acquire(context.Background(), bin, file, "p", nil, []string{"chrome", "opera"}, nil)
	if !ok || b != "opera" {
		t.Fatalf("want opera, got %q ok=%v", b, ok)
	}
	if data, _ := os.ReadFile(file); !strings.Contains(string(data), "PREMIUM") {
		t.Errorf("the Premium cookies were not kept: %q", data)
	}
	if left, _ := filepath.Glob(file + ".*"); len(left) != 0 {
		t.Errorf("temporary exports left behind: %v", left)
	}
}
