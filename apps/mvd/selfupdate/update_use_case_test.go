package selfupdate

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/question"
)

func zipOf(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte(content))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func digestOf(b []byte) string {
	sum := sha256.Sum256(b)

	return "sha256:" + hex.EncodeToString(sum[:])
}

type updateRig struct {
	checker *Checker
	exe     string
	opened  []string
	quits   int
}

// newUpdateRig serves a release 0.7.0 whose zip holds a program with newContent, stated with
// digest (the digest of the zip when empty), and installs a program "old" in a folder of its own.
func newUpdateRig(t *testing.T, packageName, digest string, archive []byte) *updateRig {
	t.Helper()
	rig := &updateRig{}
	if digest == "" {
		digest = digestOf(archive)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".zip") {
			_, _ = w.Write(archive)

			return
		}
		_, _ = fmt.Fprintf(w, `[{"tag_name":"mvd@0.7.0","html_url":"https://example.test/mvd-0.7.0","assets":[{"name":%q,"browser_download_url":%q,"digest":%q}]}]`,
			packageName, server.URL+"/"+packageName, digest)
	}))
	t.Cleanup(server.Close)

	installDir := t.TempDir()
	rig.exe = filepath.Join(installDir, "mvd.exe")
	if err := os.WriteFile(rig.exe, []byte("old"), 0o700); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	rig.checker = &Checker{
		Dir: t.TempDir(), Version: "0.6.1",
		Notify: func(notification.Notice) {},
		Ask:    func(string, string, []string) question.Answer { return question.AnswerFirst },
		Open:   func(url string) error { rig.opened = append(rig.opened, url); return nil },
		Logf:   func(string, ...interface{}) {},
		Quit:   func() { rig.quits++ },
		URL:    server.URL, Now: func() time.Time { return now },
		executable: func() (string, error) { return rig.exe, nil },
		goos:       "windows", goarch: "amd64", insecure: true,
	}

	return rig
}

func content(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func TestUpdateNowReplacesTheProgramAndAsksForARelaunch(t *testing.T) {
	rig := newUpdateRig(t, "mvd_0.7.0_windows_amd64.zip", "", zipOf(t, "mvd.exe", "new"))
	line := rig.checker.CheckNow(t.Context())
	if got := content(t, rig.exe); got != "new" {
		t.Fatalf("the program is %q, want the new one (line: %s)", got, line)
	}
	if got := content(t, rig.exe+AsideSuffix); got != "old" {
		t.Errorf("the old program waits as %q, want old", got)
	}
	if rig.checker.Relaunch() != rig.exe || rig.quits != 1 || len(rig.opened) != 0 {
		t.Errorf("relaunch = %q, quits = %d, opened = %v", rig.checker.Relaunch(), rig.quits, rig.opened)
	}
	if entries, _ := os.ReadDir(rig.checker.Dir); len(entries) != 1 { // only update-check.json
		t.Errorf("the working files were left behind: %v", entries)
	}
}

func TestUpdateNowChangesNothingWhenTheChecksumFails(t *testing.T) {
	cases := map[string]string{
		"wrong":   digestOf([]byte("something else")),
		"missing": "-",
	}
	for name, digest := range cases {
		t.Run(name, func(t *testing.T) {
			rig := newUpdateRig(t, "mvd_0.7.0_windows_amd64.zip", digest, zipOf(t, "mvd.exe", "evil"))
			line := rig.checker.CheckNow(t.Context())
			if got := content(t, rig.exe); got != "old" {
				t.Errorf("the program became %q", got)
			}
			if _, err := os.Stat(rig.exe + AsideSuffix); err == nil {
				t.Error("the old program was moved aside")
			}
			if rig.quits != 0 || rig.checker.Relaunch() != "" {
				t.Errorf("quits = %d, relaunch = %q", rig.quits, rig.checker.Relaunch())
			}
			if len(rig.opened) != 1 || !strings.Contains(line, "could not update itself") {
				t.Errorf("opened = %v, line = %q; want the release page and a clear message", rig.opened, line)
			}
		})
	}
}

func TestUpdateNowFallsBackToThePageOnOtherSystemsAndMissingFiles(t *testing.T) {
	rig := newUpdateRig(t, "mvd_0.7.0_windows_amd64.zip", "", zipOf(t, "mvd.exe", "new"))
	rig.checker.goos, rig.checker.goarch = "darwin", "arm64"
	line := rig.checker.CheckNow(t.Context())
	if content(t, rig.exe) != "old" || len(rig.opened) != 1 || !strings.Contains(line, "only does that on Windows") {
		t.Errorf("opened = %v, line = %q", rig.opened, line)
	}

	other := newUpdateRig(t, "mvd_0.7.0_linux_amd64.zip", "", zipOf(t, "mvd.exe", "new"))
	line = other.checker.CheckNow(t.Context())
	if content(t, other.exe) != "old" || len(other.opened) != 1 || !strings.Contains(line, "is not in the release") {
		t.Errorf("opened = %v, line = %q", other.opened, line)
	}
}

func TestUpdateNowRefusesAZipWithoutTheProgram(t *testing.T) {
	rig := newUpdateRig(t, "mvd_0.7.0_windows_amd64.zip", "", zipOf(t, "other.exe", "new"))
	line := rig.checker.CheckNow(t.Context())
	if content(t, rig.exe) != "old" || len(rig.opened) != 1 || !strings.Contains(line, "is not in the download") {
		t.Errorf("opened = %v, line = %q", rig.opened, line)
	}
}

func TestUpdateNowRefusesAnAddressThatIsNotHTTPS(t *testing.T) {
	rig := newUpdateRig(t, "mvd_0.7.0_windows_amd64.zip", "", zipOf(t, "mvd.exe", "new"))
	rig.checker.insecure = false
	line := rig.checker.CheckNow(t.Context())
	if content(t, rig.exe) != "old" || !strings.Contains(line, "not an https address") {
		t.Errorf("line = %q", line)
	}
}

func TestReplaceProgramLeavesTheInstallAloneWhenItCannotRun(t *testing.T) {
	dir := t.TempDir()
	newProgram := filepath.Join(dir, "new")
	if err := os.WriteFile(newProgram, []byte("new"), 0o700); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "gone", "mvd.exe") // its folder is not there: nothing can be written
	if err := replaceProgram(exe, newProgram); err == nil {
		t.Fatal("replaced a program that is not there")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("files were left behind: %v", entries)
	}

	// The running program itself is missing but its folder is writable: the staged copy goes.
	exe = filepath.Join(dir, "mvd.exe")
	if err := replaceProgram(exe, newProgram); err == nil {
		t.Fatal("replaced a program that is not there")
	}
	if _, err := os.Stat(exe + ".new"); err == nil {
		t.Error("the staged copy was left behind")
	}
}

func TestDownloadCheckedRemovesWhatDoesNotMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("abc")) }))
	t.Cleanup(server.Close)
	dest := filepath.Join(t.TempDir(), "file")
	if err := downloadChecked(t.Context(), server.Client(), server.URL, digestOf([]byte("abd")), dest); err == nil {
		t.Fatal("accepted a file with the wrong checksum")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("the file was kept")
	}
	if err := downloadChecked(t.Context(), server.Client(), server.URL, digestOf([]byte("abc")), dest); err != nil {
		t.Errorf("rejected a file with the right checksum: %v", err)
	}
}
