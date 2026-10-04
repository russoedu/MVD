package deps

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// exe gives a program its file name on the machine running the tests.
func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}

	return name
}

// testDeps is a yt-dlp that is always updated, an ffmpeg that is replaced only when a
// month newer, and a deno that is fetched once, whatever the operating system.
func testDeps() map[string]Dep {
	return map[string]Dep{
		"yt-dlp": {Name: "yt-dlp", URL: "https://example.test/latest/yt-dlp", FileName: exe("yt-dlp"), Source: &Source{Repo: "o/yt", Asset: "yt-dlp"}},
		"ffmpeg": {Name: "ffmpeg", URL: "https://example.test/latest/ffmpeg.zip", IsZip: true, FileName: exe("ffmpeg"), Source: &Source{Repo: "o/ff", Asset: "ffmpeg.zip"}, MinAge: 30 * 24 * time.Hour},
		"deno":   {Name: "deno", URL: "https://example.test/deno.zip", IsZip: true, FileName: exe("deno")},
	}
}

func zipOf(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("archive/bin/" + name)
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

func sumOf(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

// world is the internet and GitHub as the installer sees them.
type world struct {
	t        *testing.T
	served   map[string][]byte    // address -> what is served there
	releases map[string]AssetInfo // asset name -> the latest release's entry for it
	fetched  []string
	// lookupErr, when set, is what asking GitHub returns.
	lookupErr error
	// fetchErr maps an address to the error fetching it returns.
	fetchErr map[string]error
}

var epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{t: t, served: map[string][]byte{}, releases: map[string]AssetInfo{}, fetchErr: map[string]error{}}
	w.publishYtDlp(1, "2026.01.01", epoch, "yt-dlp build 1")
	w.publishFfmpeg(10, epoch, "ffmpeg build 1")
	w.served["https://example.test/deno.zip"] = zipOf(t, exe("deno"), "deno build")

	return w
}

func (w *world) publish(asset, url string, id int64, tag string, at time.Time, data []byte) {
	w.served[url] = data
	w.releases[asset] = AssetInfo{ID: id, Tag: tag, URL: url, SHA256: sumOf(data), UpdatedAt: at}
}

func (w *world) publishYtDlp(id int64, tag string, at time.Time, content string) {
	w.publish("yt-dlp", "https://example.test/asset/yt-dlp/"+tag, id, tag, at, []byte(content))
}

func (w *world) publishFfmpeg(id int64, at time.Time, content string) {
	w.publish("ffmpeg.zip", "https://example.test/asset/ffmpeg/"+at.Format("20060102"), id, "latest", at, zipOf(w.t, exe("ffmpeg"), content))
}

func (w *world) fetch(url, dest string) (string, error) {
	w.fetched = append(w.fetched, url)
	if err := w.fetchErr[url]; err != nil {
		return "", err
	}
	data, ok := w.served[url]
	if !ok {
		// The address that always serves the newest build.
		for _, release := range w.releases {
			if strings.HasSuffix(url, "/latest/yt-dlp") && strings.Contains(release.URL, "/yt-dlp/") ||
				strings.HasSuffix(url, "/latest/ffmpeg.zip") && strings.Contains(release.URL, "/ffmpeg/") {
				data, ok = w.served[release.URL]
			}
		}
	}
	if !ok {
		return "", errors.New("HTTP status 404 Not Found")
	}
	if err := os.WriteFile(dest, data, 0o600); err != nil {
		return "", err
	}

	return sumOf(data), nil
}

func (w *world) lookup(source Source) (AssetInfo, error) {
	if w.lookupErr != nil {
		return AssetInfo{}, w.lookupErr
	}
	if info, ok := w.releases[source.Asset]; ok {
		return info, nil
	}

	return AssetInfo{}, errors.New("no such asset")
}

type recorder struct{ events []Event }

func (r *recorder) report(e Event) { r.events = append(r.events, e) }

var kindNames = map[EventKind]string{
	EventMissing: "missing", EventDownloading: "downloading", EventInstalled: "installed", EventFailed: "failed",
	EventUpToDate: "uptodate", EventUpdated: "updated", EventUpdateFailed: "updatefailed",
}

func (r *recorder) kinds() string {
	var out []string
	for _, e := range r.events {
		out = append(out, kindNames[e.Kind]+":"+e.Name)
	}

	return strings.Join(out, " ")
}

func (r *recorder) find(kind EventKind, name string) (Event, bool) {
	for _, e := range r.events {
		if e.Kind == kind && e.Name == name {
			return e, true
		}
	}

	return Event{}, false
}

// newTestInstaller returns an installer over the fake world, whose PATH has nothing in
// it, and where the named programs are "installed on the system" if present says so.
func newTestInstaller(t *testing.T, binDir string, w *world, present map[string]bool) (*installer, *recorder) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	rec := &recorder{}

	return &installer{
		binDir: binDir, goos: runtime.GOOS, goarch: runtime.GOARCH, available: testDeps(), owned: true,
		lookPath: func(name string) (string, error) {
			if present[name] {
				return name, nil
			}

			return "", errors.New("not found")
		},
		lookup: w.lookup, fetch: w.fetch, report: rec.report,
	}, rec
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, entry := range entries {
		out = append(out, entry.Name())
	}

	return out
}

func listing(t *testing.T, dir string) string {
	t.Helper()

	return strings.Join(names(t, dir), " ")
}

func binPath(dir, name string) string { return filepath.Join(dir, exe(name)) }
