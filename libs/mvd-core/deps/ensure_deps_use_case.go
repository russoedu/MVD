package deps

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// managedTools are the tools that can be kept up to date, in the order they are handled.
var managedTools = []string{"yt-dlp", "ffmpeg"}

// Ensure is what the terminal app calls: tools live in ./bin next to where it was
// started, a tool is fetched only when nothing on PATH provides it, and progress is
// printed. It never replaces a tool that is already there.
func Ensure() {
	i := newInstaller("./bin", PrintProgress)
	i.owned = false
	i.ensure()
}

// EnsureIn is what the tray app calls at start-up. It puts binDir on PATH and makes sure
// the app has its own copy of yt-dlp and ffmpeg there, whatever is on PATH, and a
// JavaScript runtime (deno or node) from somewhere. It downloads whatever is missing and
// tells report what it is doing.
//
// binDir should be a folder the person can always write to, and the same one on every
// start, or the tools are fetched again each time.
func EnsureIn(binDir string, report Reporter) {
	newInstaller(binDir, report).ensure()
}

// UpdateIn checks whether newer builds of yt-dlp and ffmpeg have been published and
// replaces the app's copy of each that is out of date. It returns the names of the ones
// it replaced. A check that fails (no connection, GitHub limiting requests) is reported
// and leaves the copy that is there untouched, and so does an update that fails
// half way, so a failed update never takes a working tool away.
func UpdateIn(binDir string, report Reporter) []string {
	return newInstaller(binDir, report).update()
}

// installer carries what ensure and update need, so a test can replace the network,
// the PATH lookup and the list of tools.
type installer struct {
	binDir    string
	goos      string
	goarch    string
	available map[string]Dep
	// owned makes yt-dlp and ffmpeg the app's own copies rather than whatever PATH has.
	owned    bool
	lookPath func(string) (string, error)
	lookup   func(Source) (AssetInfo, error)
	fetch    func(url, dest string) (sha256 string, err error)
	report   Reporter
}

func newInstaller(binDir string, report Reporter) *installer {
	if report == nil {
		report = func(Event) {}
	}

	return &installer{
		binDir: binDir, goos: runtime.GOOS, goarch: runtime.GOARCH,
		available: platformDeps(runtime.GOOS, runtime.GOARCH), owned: true,
		lookPath: exec.LookPath, lookup: lookupLatest, fetch: downloadFile, report: report,
	}
}

func (i *installer) dir() string {
	dir, err := filepath.Abs(i.binDir)
	if err != nil {
		return i.binDir
	}

	return dir
}

func (i *installer) ensure() {
	i.putBinOnPath()
	dir := i.dir()
	removeLeftovers(dir)
	known := loadManifest(dir)

	var missing []string
	for _, name := range managedTools {
		if i.needsInstall(i.available[name], dir, known) {
			missing = append(missing, name)
		}
	}
	_, denoErr := i.lookPath("deno")
	_, nodeErr := i.lookPath("node")
	if denoErr != nil && nodeErr != nil {
		missing = append(missing, "deno")
	}
	if len(missing) == 0 {
		return
	}

	i.report(Event{Kind: EventMissing, Names: missing, Dir: dir})

	if err := os.MkdirAll(dir, 0o755); err != nil {
		for _, name := range missing {
			i.report(Event{Kind: EventFailed, Name: name, Dir: dir, Err: fmt.Errorf("cannot create %s: %w", dir, err)})
		}

		return
	}

	for _, name := range missing {
		dep, ok := i.available[name]
		if !ok {
			continue
		}
		// Asking GitHub first gives the build's identity and checksum. If it cannot be
		// asked, the tool is still fetched, from the address that serves the newest build.
		var info *AssetInfo
		if dep.Source != nil && i.owned {
			if found, err := i.lookup(*dep.Source); err == nil {
				info = &found
			}
		}
		if err := i.install(dep, dir, known, info); err != nil {
			i.report(Event{Kind: EventFailed, Name: dep.Name, URL: dep.URL, Dir: dir, Err: err})

			continue
		}
		i.report(Event{Kind: EventInstalled, Name: dep.Name, URL: dep.URL, Dir: dir})
	}

	i.putBinOnPath()
}

// needsInstall says whether a tool has to be fetched now.
func (i *installer) needsInstall(dep Dep, dir string, known manifest) bool {
	if dep.Source != nil && i.owned {
		return !ownsCopy(dep, dir, known)
	}
	_, err := i.lookPath(dep.Name)

	return err != nil
}

// ownsCopy reports whether the app installed this tool into dir and it is still there.
func ownsCopy(dep Dep, dir string, known manifest) bool {
	if _, installed := known[dep.Name]; !installed {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, dep.FileName))

	return err == nil
}

func (i *installer) update() []string {
	i.putBinOnPath()
	dir := i.dir()
	known := loadManifest(dir)

	var updated []string
	for _, name := range managedTools {
		dep, ok := i.available[name]
		if !ok || dep.Source == nil {
			continue
		}

		info, err := i.lookup(*dep.Source)
		if err != nil {
			i.report(Event{Kind: EventUpdateFailed, Name: dep.Name, Dir: dir, Err: err})

			continue
		}

		have := known[dep.Name]
		owned := ownsCopy(dep, dir, known)
		switch {
		case owned && have.AssetID == info.ID:
			i.report(Event{Kind: EventUpToDate, Name: dep.Name, Detail: describeBuild(have.Tag, have.UpdatedAt)})

			continue
		case owned && dep.MinAge > 0 && info.UpdatedAt.Sub(have.UpdatedAt) < dep.MinAge:
			i.report(Event{
				Kind: EventUpToDate, Name: dep.Name,
				Detail: fmt.Sprintf("%s; a newer build exists, but %s is only replaced by one at least %d days newer",
					describeBuild(have.Tag, have.UpdatedAt), dep.Name, int(dep.MinAge.Hours()/24)),
			})

			continue
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			i.report(Event{Kind: EventUpdateFailed, Name: dep.Name, Dir: dir, Err: err})

			continue
		}
		if err := i.install(dep, dir, known, &info); err != nil {
			i.report(Event{Kind: EventUpdateFailed, Name: dep.Name, URL: info.URL, Dir: dir, Err: err})

			continue
		}
		if owned {
			updated = append(updated, dep.Name)
			i.report(Event{Kind: EventUpdated, Name: dep.Name, URL: info.URL, Dir: dir, Detail: describeBuild(info.Tag, info.UpdatedAt)})
		} else {
			i.report(Event{Kind: EventInstalled, Name: dep.Name, URL: info.URL, Dir: dir})
		}
	}

	return updated
}

// describeBuild names a build by its version, or by its date when its release is
// always called "latest".
func describeBuild(tag string, updatedAt time.Time) string {
	if tag == "" || tag == "latest" {
		return "built " + updatedAt.Format("2006-01-02")
	}

	return tag
}

// install downloads one tool and puts it in dir, checking it against the checksum
// GitHub published when there is one. The download and the file that is put in place
// both go under temporary names first, so an interrupted or failed run never leaves a
// half-written program where a later start would find and run it, and never takes the
// working copy away: replaceFile only swaps in a file that is complete.
func (i *installer) install(dep Dep, dir string, known manifest, info *AssetInfo) error {
	url := dep.URL
	if info != nil && info.URL != "" {
		url = info.URL
	}
	final := filepath.Join(dir, dep.FileName)
	download := final + ".download"
	partial := final + ".part"
	defer func() {
		_ = os.Remove(download)
		_ = os.Remove(partial)
	}()

	i.report(Event{Kind: EventDownloading, Name: dep.Name, URL: url, Dir: dir})

	sum, err := i.fetch(url, download)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	if info != nil && info.SHA256 != "" && sum != info.SHA256 {
		return fmt.Errorf("the download of %s does not match the checksum GitHub publishes for it, so it was not installed", dep.Name)
	}

	if dep.IsZip {
		err = extractZipFile(download, dep.FileName, partial)
	} else {
		err = os.Rename(download, partial)
	}
	if err != nil {
		return fmt.Errorf("cannot unpack %s: %w", dep.FileName, err)
	}
	if err := os.Chmod(partial, 0o755); err != nil {
		return fmt.Errorf("cannot make %s executable: %w", dep.FileName, err)
	}
	if err := replaceFile(partial, final); err != nil {
		return fmt.Errorf("cannot install %s: %w", dep.FileName, err)
	}

	if dep.Source != nil {
		record := installedTool{SHA256: sum}
		if info != nil {
			record.AssetID, record.Tag, record.UpdatedAt = info.ID, info.Tag, info.UpdatedAt
		}
		known[dep.Name] = record
		_ = saveManifest(dir, known)
	}

	return nil
}

// removeLeftovers deletes what an interrupted run or an earlier replacement left in dir.
func removeLeftovers(dir string) {
	for _, pattern := range []string{"*.part", "*.download", "*.old"} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			_ = os.Remove(match)
		}
	}
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

// extractZipFile writes the first entry of the archive at zipPath named targetFileName
// to destPath.
func extractZipFile(zipPath, targetFileName, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

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
