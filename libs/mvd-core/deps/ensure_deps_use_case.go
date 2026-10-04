package deps

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Ensure is what the terminal app calls: tools live in ./bin next to where it was
// started, and progress is printed.
func Ensure() {
	EnsureIn("./bin", PrintProgress)
}

// EnsureIn puts binDir on PATH, checks for yt-dlp, ffmpeg and a JavaScript runtime
// (deno or node), and downloads whatever is missing into binDir, telling report what
// it is doing. binDir should be a folder the person can always write to, and the same
// one on every start, or the tools are fetched again each time.
func EnsureIn(binDir string, report Reporter) {
	newInstaller(binDir, report).ensure()
}

// installer carries what ensure needs, so a test can replace the network and PATH lookup.
type installer struct {
	binDir   string
	goos     string
	goarch   string
	lookPath func(string) (string, error)
	fetch    func(url string) ([]byte, error)
	report   Reporter
}

func newInstaller(binDir string, report Reporter) *installer {
	if report == nil {
		report = func(Event) {}
	}

	return &installer{
		binDir: binDir, goos: runtime.GOOS, goarch: runtime.GOARCH,
		lookPath: exec.LookPath, fetch: downloadBytes, report: report,
	}
}

func (i *installer) ensure() {
	i.putBinOnPath()

	var missing []string
	if _, err := i.lookPath("yt-dlp"); err != nil {
		missing = append(missing, "yt-dlp")
	}
	if _, err := i.lookPath("ffmpeg"); err != nil {
		missing = append(missing, "ffmpeg")
	}
	_, denoErr := i.lookPath("deno")
	_, nodeErr := i.lookPath("node")
	if denoErr != nil && nodeErr != nil {
		missing = append(missing, "deno")
	}
	if len(missing) == 0 {
		return
	}

	dir, err := filepath.Abs(i.binDir)
	if err != nil {
		dir = i.binDir
	}
	i.report(Event{Kind: EventMissing, Names: missing, Dir: dir})

	if err := os.MkdirAll(dir, 0o755); err != nil {
		for _, name := range missing {
			i.report(Event{Kind: EventFailed, Name: name, Dir: dir, Err: fmt.Errorf("cannot create %s: %w", dir, err)})
		}

		return
	}

	available := platformDeps(i.goos, i.goarch)
	for _, name := range missing {
		if dep, ok := available[name]; ok {
			i.install(dep, dir)
		}
	}

	i.putBinOnPath()
}

// install downloads one tool and puts it in dir. The file is written under a temporary
// name and renamed, so an interrupted run never leaves a half-written program that a
// later start would find on PATH and try to run.
func (i *installer) install(dep Dep, dir string) {
	fail := func(err error) {
		i.report(Event{Kind: EventFailed, Name: dep.Name, URL: dep.URL, Dir: dir, Err: err})
	}
	i.report(Event{Kind: EventDownloading, Name: dep.Name, URL: dep.URL, Dir: dir})

	data, err := i.fetch(dep.URL)
	if err != nil {
		fail(fmt.Errorf("download failed: %w", err))

		return
	}

	final := filepath.Join(dir, dep.FileName)
	partial := final + ".part"
	if dep.IsZip {
		err = extractZipFile(data, dep.FileName, partial)
	} else {
		err = os.WriteFile(partial, data, 0o755)
	}
	if err != nil {
		_ = os.Remove(partial)
		fail(fmt.Errorf("cannot write %s: %w", dep.FileName, err))

		return
	}
	if err := os.Chmod(partial, 0o755); err != nil {
		_ = os.Remove(partial)
		fail(fmt.Errorf("cannot make %s executable: %w", dep.FileName, err))

		return
	}
	if err := os.Rename(partial, final); err != nil {
		_ = os.Remove(partial)
		fail(fmt.Errorf("cannot install %s: %w", dep.FileName, err))

		return
	}

	i.report(Event{Kind: EventInstalled, Name: dep.Name, URL: dep.URL, Dir: dir})
}

// putBinOnPath puts binDir and binDir/<os> at the front of PATH.
func (i *installer) putBinOnPath() {
	absBin, err := filepath.Abs(i.binDir)
	if err != nil {
		return
	}
	separator := string(os.PathListSeparator)
	newPath := absBin + separator + filepath.Join(absBin, i.goos) + separator + os.Getenv("PATH")
	if err := os.Setenv("PATH", newPath); err != nil {
		i.report(Event{Kind: EventFailed, Name: "PATH", Dir: absBin, Err: err})
	}
}

// extractZipFile writes the first entry of the archive named targetFileName
// to destPath.
func extractZipFile(data []byte, targetFileName, destPath string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if strings.EqualFold(baseName, targetFileName) || strings.EqualFold(f.Name, targetFileName) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer func() { _ = rc.Close() }()

			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, rc)
			if closeErr := out.Close(); err == nil {
				err = closeErr
			}
			return err
		}
	}

	return fmt.Errorf("file '%s' not found inside zip archive", targetFileName)
}
