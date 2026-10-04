package deps

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func zipOf(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("archive/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// serveAll answers every tool's URL with a program (or an archive holding one).
func serveAll(t *testing.T) func(string) ([]byte, error) {
	t.Helper()
	byURL := map[string][]byte{}
	for _, dep := range platformDeps(runtime.GOOS, runtime.GOARCH) {
		if dep.IsZip {
			byURL[dep.URL] = zipOf(t, dep.FileName, "program "+dep.Name)
		} else {
			byURL[dep.URL] = []byte("program " + dep.Name)
		}
	}

	return func(url string) ([]byte, error) {
		if data, ok := byURL[url]; ok {
			return data, nil
		}

		return nil, errors.New("unexpected url " + url)
	}
}

type recorder struct{ events []Event }

func (r *recorder) report(e Event) { r.events = append(r.events, e) }

func (r *recorder) kinds() string {
	var out []string
	for _, e := range r.events {
		out = append(out, map[EventKind]string{EventMissing: "missing", EventDownloading: "downloading", EventInstalled: "installed", EventFailed: "failed"}[e.Kind]+":"+e.Name)
	}

	return strings.Join(out, " ")
}

func newTestInstaller(t *testing.T, binDir string, present map[string]bool, fetch func(string) ([]byte, error)) (*installer, *recorder) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	rec := &recorder{}

	return &installer{
		binDir: binDir, goos: runtime.GOOS, goarch: runtime.GOARCH,
		lookPath: func(name string) (string, error) {
			if present[name] {
				return name, nil
			}

			return "", errors.New("not found")
		},
		fetch: fetch, report: rec.report,
	}, rec
}

func TestNothingIsFetchedWhenEverythingIsInstalled(t *testing.T) {
	fetched := false
	i, rec := newTestInstaller(t, t.TempDir(), map[string]bool{"yt-dlp": true, "ffmpeg": true, "deno": true},
		func(string) ([]byte, error) { fetched = true; return nil, nil })

	i.ensure()

	if fetched || len(rec.events) != 0 {
		t.Errorf("fetched=%v events=%v", fetched, rec.kinds())
	}
}

func TestNodeCountsAsAJavaScriptRuntime(t *testing.T) {
	i, rec := newTestInstaller(t, t.TempDir(), map[string]bool{"yt-dlp": true, "ffmpeg": true, "node": true},
		func(string) ([]byte, error) { t.Fatal("fetched although node is installed"); return nil, nil })

	i.ensure()

	if len(rec.events) != 0 {
		t.Errorf("events = %s", rec.kinds())
	}
}

func TestMissingToolsAreAnnouncedThenDownloadedAndInstalledWithNoTemporaryFileLeft(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mvd", "bin")
	i, rec := newTestInstaller(t, dir, nil, serveAll(t))

	i.ensure()

	want := "missing: downloading:yt-dlp installed:yt-dlp downloading:ffmpeg installed:ffmpeg downloading:deno installed:deno"
	if got := rec.kinds(); got != want {
		t.Fatalf("events:\n got  %s\n want %s", got, want)
	}
	first := rec.events[0]
	if len(first.Names) != 3 || first.Dir == "" || !filepath.IsAbs(first.Dir) {
		t.Errorf("the announcement should name three tools and an absolute folder: %+v", first)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 3 {
		t.Fatalf("entries = %v (%v)", entries, err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".part") {
			t.Errorf("temporary file left behind: %s", entry.Name())
		}
	}
}

func TestOneFailedDownloadIsReportedAndTheRestStillInstall(t *testing.T) {
	serve := serveAll(t)
	ffmpegURL := platformDeps(runtime.GOOS, runtime.GOARCH)["ffmpeg"].URL
	i, rec := newTestInstaller(t, filepath.Join(t.TempDir(), "bin"), nil, func(url string) ([]byte, error) {
		if url == ffmpegURL {
			return nil, errors.New("HTTP status 503")
		}

		return serve(url)
	})

	i.ensure()

	want := "missing: downloading:yt-dlp installed:yt-dlp downloading:ffmpeg failed:ffmpeg downloading:deno installed:deno"
	if got := rec.kinds(); got != want {
		t.Fatalf("events:\n got  %s\n want %s", got, want)
	}
	for _, e := range rec.events {
		if e.Kind == EventFailed && (e.Err == nil || !strings.Contains(e.Err.Error(), "503")) {
			t.Errorf("the failure should say why: %+v", e)
		}
	}
}

func TestAFolderThatCannotBeCreatedFailsEveryToolWithoutDownloading(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	i, rec := newTestInstaller(t, filepath.Join(blocker, "bin"), nil,
		func(string) ([]byte, error) { t.Fatal("downloaded into a folder that cannot exist"); return nil, nil })

	i.ensure()

	if got := rec.kinds(); got != "missing: failed:yt-dlp failed:ffmpeg failed:deno" {
		t.Errorf("events = %s", got)
	}
}

func TestToolsInstalledOnceAreFoundAgainWhenTheAppStartsFromAnotherFolder(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "appdata", "bin")
	first := &installer{binDir: bin, goos: runtime.GOOS, goarch: runtime.GOARCH, lookPath: realLookPath, fetch: serveAll(t), report: func(Event) {}}
	t.Setenv("PATH", t.TempDir())

	first.ensure()

	// A later start: another working directory, and a fresh PATH without the first run's edit.
	t.Chdir(t.TempDir())
	t.Setenv("PATH", t.TempDir())
	rec := &recorder{}
	second := &installer{binDir: bin, goos: runtime.GOOS, goarch: runtime.GOARCH, lookPath: realLookPath,
		fetch: func(string) ([]byte, error) { t.Fatal("fetched again"); return nil, nil }, report: rec.report}

	second.ensure()

	if len(rec.events) != 0 {
		t.Errorf("the second start reported %s", rec.kinds())
	}
}

// realLookPath is exec.LookPath, named so the test reads as "the real PATH search".
var realLookPath = exec.LookPath
