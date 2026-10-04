package deps

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// --- first run -------------------------------------------------------------

func TestTheFirstRunAnnouncesThenDownloadsAndInstallsEveryToolWithNothingLeftOver(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mvd", "bin")
	i, rec := newTestInstaller(t, dir, newWorld(t), nil)

	i.ensure()

	want := "missing: downloading:yt-dlp installed:yt-dlp downloading:ffmpeg installed:ffmpeg downloading:deno installed:deno"
	if got := rec.kinds(); got != want {
		t.Fatalf("events:\n got  %s\n want %s", got, want)
	}
	first := rec.events[0]
	if len(first.Names) != 3 || !filepath.IsAbs(first.Dir) {
		t.Errorf("the announcement should name three tools and an absolute folder: %+v", first)
	}
	expected := []string{exe("deno"), exe("ffmpeg"), exe("yt-dlp"), "tools.json"}
	sort.Strings(expected)
	if got, want := listing(t, dir), strings.Join(expected, " "); got != want {
		t.Errorf("folder holds %q, want %q (no .part, .download or .old files)", got, want)
	}
	if read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 1" || read(t, binPath(dir, "ffmpeg")) != "ffmpeg build 1" {
		t.Error("the installed programs are not the downloaded builds")
	}
}

func TestTheFirstRunRecordsWhichBuildWasInstalledForTheToolsItKeepsCurrent(t *testing.T) {
	dir := t.TempDir()
	i, _ := newTestInstaller(t, dir, newWorld(t), nil)

	i.ensure()

	m := loadManifest(dir)
	if m["yt-dlp"].AssetID != 1 || m["yt-dlp"].Tag != "2026.01.01" || m["yt-dlp"].SHA256 == "" {
		t.Errorf("yt-dlp entry = %+v", m["yt-dlp"])
	}
	if m["ffmpeg"].AssetID != 10 {
		t.Errorf("ffmpeg entry = %+v", m["ffmpeg"])
	}
	if _, recorded := m["deno"]; recorded {
		t.Error("deno is fetched once and never updated, so it needs no entry")
	}
}

func TestTheAppKeepsItsOwnYtDlpAndFfmpegEvenWhenThePathHasThem(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	i, rec := newTestInstaller(t, dir, w, map[string]bool{"yt-dlp": true, "ffmpeg": true, "deno": true})

	i.ensure()

	if got := rec.kinds(); got != "missing: downloading:yt-dlp installed:yt-dlp downloading:ffmpeg installed:ffmpeg" {
		t.Errorf("events = %s", got)
	}
	if first := rec.events[0]; len(first.Names) != 2 {
		t.Errorf("only the two tools it owns should be announced: %+v", first.Names)
	}
}

func TestTheTerminalAppOnlyFetchesWhatNothingOnThePathProvides(t *testing.T) {
	w := newWorld(t)
	i, rec := newTestInstaller(t, t.TempDir(), w, map[string]bool{"yt-dlp": true, "ffmpeg": true, "deno": true})
	i.owned = false

	i.ensure()

	if len(rec.events) != 0 || len(w.fetched) != 0 {
		t.Errorf("events=%s fetched=%v", rec.kinds(), w.fetched)
	}
}

func TestNodeCountsAsAJavaScriptRuntime(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	i, rec := newTestInstaller(t, dir, w, map[string]bool{"node": true})

	i.ensure()

	if _, fetchedDeno := rec.find(EventInstalled, "deno"); fetchedDeno {
		t.Errorf("fetched deno although node is installed: %s", rec.kinds())
	}
}

func TestOnceInstalledNothingIsFetchedAtTheNextStart(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	first, _ := newTestInstaller(t, dir, w, nil)
	first.ensure()
	w.fetched = nil

	second, rec := newTestInstaller(t, dir, w, map[string]bool{"deno": true})
	second.ensure()

	if len(rec.events) != 0 || len(w.fetched) != 0 {
		t.Errorf("events=%s fetched=%v", rec.kinds(), w.fetched)
	}
}

func TestAToolWhoseFileWasDeletedIsFetchedAgain(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	first, _ := newTestInstaller(t, dir, w, nil)
	first.ensure()
	if err := os.Remove(binPath(dir, "yt-dlp")); err != nil {
		t.Fatal(err)
	}

	second, rec := newTestInstaller(t, dir, w, map[string]bool{"deno": true})
	second.ensure()

	if got := rec.kinds(); got != "missing: downloading:yt-dlp installed:yt-dlp" {
		t.Errorf("events = %s", got)
	}
}

func TestAFileTheAppDidNotInstallIsNotTreatedAsItsOwn(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(binPath(dir, "yt-dlp"), []byte("someone else's"), 0o755); err != nil {
		t.Fatal(err)
	}
	i, rec := newTestInstaller(t, dir, newWorld(t), map[string]bool{"deno": true})

	i.ensure()

	if _, fetched := rec.find(EventInstalled, "yt-dlp"); !fetched {
		t.Errorf("a yt-dlp with no record should be replaced by the app's own: %s", rec.kinds())
	}
}

func TestLeftoversOfAnInterruptedRunAreRemovedAtStart(t *testing.T) {
	dir := t.TempDir()
	for _, stale := range []string{"yt-dlp.part", "ffmpeg.download", "yt-dlp.old"} {
		if err := os.WriteFile(filepath.Join(dir, stale), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	i, _ := newTestInstaller(t, dir, newWorld(t), map[string]bool{"deno": true})

	i.ensure()

	for _, name := range names(t, dir) {
		if strings.HasSuffix(name, ".part") || strings.HasSuffix(name, ".download") || strings.HasSuffix(name, ".old") {
			t.Errorf("leftover not removed: %s", name)
		}
	}
}

func TestADownloadThatDoesNotMatchItsPublishedChecksumIsNotInstalled(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	tampered := w.releases["yt-dlp"]
	tampered.SHA256 = strings.Repeat("0", 64)
	w.releases["yt-dlp"] = tampered
	i, rec := newTestInstaller(t, dir, w, map[string]bool{"deno": true})

	i.ensure()

	failed, ok := rec.find(EventFailed, "yt-dlp")
	if !ok || !strings.Contains(failed.Err.Error(), "checksum") {
		t.Fatalf("expected a checksum failure, got %s", rec.kinds())
	}
	if _, err := os.Stat(binPath(dir, "yt-dlp")); err == nil {
		t.Error("a program that failed its checksum was installed")
	}
	if _, installed := rec.find(EventInstalled, "ffmpeg"); !installed {
		t.Error("the other tool should still install")
	}
	if got := listing(t, dir); strings.Contains(got, ".download") || strings.Contains(got, ".part") {
		t.Errorf("leftovers: %s", got)
	}
}

func TestOneFailedDownloadIsReportedAndTheRestStillInstall(t *testing.T) {
	w := newWorld(t)
	w.fetchErr[w.releases["ffmpeg.zip"].URL] = errors.New("HTTP status 503")
	i, rec := newTestInstaller(t, t.TempDir(), w, nil)

	i.ensure()

	want := "missing: downloading:yt-dlp installed:yt-dlp downloading:ffmpeg failed:ffmpeg downloading:deno installed:deno"
	if got := rec.kinds(); got != want {
		t.Fatalf("events:\n got  %s\n want %s", got, want)
	}
	if failed, _ := rec.find(EventFailed, "ffmpeg"); !strings.Contains(failed.Err.Error(), "503") {
		t.Errorf("the failure should say why: %v", failed.Err)
	}
}

func TestWhenGitHubCannotBeAskedTheToolsAreStillFetchedFromTheNewestBuildAddress(t *testing.T) {
	dir := t.TempDir()
	w := newWorld(t)
	w.lookupErr = errors.New("GitHub is limiting requests for now (HTTP 403)")
	i, rec := newTestInstaller(t, dir, w, nil)

	i.ensure()

	if _, ok := rec.find(EventInstalled, "yt-dlp"); !ok {
		t.Fatalf("yt-dlp should install without the lookup: %s", rec.kinds())
	}
	if got := loadManifest(dir)["yt-dlp"].AssetID; got != 0 {
		t.Errorf("an install made without the lookup does not know its build; AssetID = %d", got)
	}
}

func TestAFolderThatCannotBeCreatedFailsEveryToolWithoutDownloading(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	w := newWorld(t)
	i, rec := newTestInstaller(t, filepath.Join(blocker, "bin"), w, nil)

	i.ensure()

	if got := rec.kinds(); got != "missing: failed:yt-dlp failed:ffmpeg failed:deno" {
		t.Errorf("events = %s", got)
	}
	if len(w.fetched) != 0 {
		t.Errorf("downloaded into a folder that cannot exist: %v", w.fetched)
	}
}

func TestToolsInstalledOnceAreFoundAgainWhenTheAppStartsFromAnotherFolder(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "appdata", "bin")
	w := newWorld(t)
	first, _ := newTestInstaller(t, bin, w, nil)
	first.lookPath = exec.LookPath
	first.ensure()

	// A later start: another working directory and a fresh PATH without the first run's edit.
	t.Chdir(t.TempDir())
	w.fetched = nil
	second, rec := newTestInstaller(t, bin, w, nil)
	second.lookPath = exec.LookPath
	second.ensure()

	if len(rec.events) != 0 || len(w.fetched) != 0 {
		t.Errorf("the second start reported %s and fetched %v", rec.kinds(), w.fetched)
	}
}

// --- updates ---------------------------------------------------------------

// installedWorld has every tool installed from the initial builds.
func installedWorld(t *testing.T) (dir string, w *world) {
	t.Helper()
	dir = t.TempDir()
	w = newWorld(t)
	first, _ := newTestInstaller(t, dir, w, nil)
	first.ensure()
	w.fetched = nil

	return dir, w
}

func TestWhenNothingNewHasBeenPublishedNothingIsDownloaded(t *testing.T) {
	dir, w := installedWorld(t)
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 0 || len(w.fetched) != 0 {
		t.Errorf("updated=%v fetched=%v", updated, w.fetched)
	}
	if got := rec.kinds(); got != "uptodate:yt-dlp uptodate:ffmpeg" {
		t.Errorf("events = %s", got)
	}
}

func TestANewYtDlpReplacesTheOldOneAtOnceAndIsRecorded(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishYtDlp(2, "2026.02.02", epoch.Add(24*time.Hour), "yt-dlp build 2")
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 1 || updated[0] != "yt-dlp" {
		t.Fatalf("updated = %v", updated)
	}
	if read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 2" {
		t.Error("the program was not replaced")
	}
	if loaded := loadManifest(dir)["yt-dlp"]; loaded.AssetID != 2 || loaded.Tag != "2026.02.02" {
		t.Errorf("manifest = %+v", loaded)
	}
	if event, ok := rec.find(EventUpdated, "yt-dlp"); !ok || event.Detail != "2026.02.02" {
		t.Errorf("the update should say the new version: %+v (%s)", event, rec.kinds())
	}
	if got := listing(t, dir); strings.Contains(got, ".old") || strings.Contains(got, ".part") || strings.Contains(got, ".download") {
		t.Errorf("leftovers after updating: %s", got)
	}
}

func TestFfmpegIsNotReplacedByABuildLessThanThirtyDaysNewer(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishFfmpeg(11, epoch.Add(29*24*time.Hour), "ffmpeg build 2")
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 0 || read(t, binPath(dir, "ffmpeg")) != "ffmpeg build 1" {
		t.Errorf("ffmpeg was replaced after 29 days: updated=%v", updated)
	}
	if event, _ := rec.find(EventUpToDate, "ffmpeg"); !strings.Contains(event.Detail, "30 days") {
		t.Errorf("the log should explain why a newer build was left alone: %q", event.Detail)
	}
}

func TestFfmpegIsReplacedByABuildThirtyDaysNewer(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishFfmpeg(12, epoch.Add(30*24*time.Hour), "ffmpeg build 3")
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 1 || updated[0] != "ffmpeg" || read(t, binPath(dir, "ffmpeg")) != "ffmpeg build 3" {
		t.Errorf("updated=%v content=%q", updated, read(t, binPath(dir, "ffmpeg")))
	}
	if event, _ := rec.find(EventUpdated, "ffmpeg"); !strings.HasPrefix(event.Detail, "built ") {
		t.Errorf("a build called latest should be described by its date: %q", event.Detail)
	}
}

func TestWhenGitHubCannotBeAskedTheToolsAreLeftAloneAndTheFailureIsReported(t *testing.T) {
	dir, w := installedWorld(t)
	w.lookupErr = errors.New("GitHub is limiting requests for now (HTTP 403)")
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 0 || len(w.fetched) != 0 {
		t.Errorf("updated=%v fetched=%v", updated, w.fetched)
	}
	if got := rec.kinds(); got != "updatefailed:yt-dlp updatefailed:ffmpeg" {
		t.Errorf("events = %s", got)
	}
	if read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 1" {
		t.Error("a failed check changed the tool")
	}
}

func TestAFailedUpdateDownloadNeverTakesTheWorkingToolAway(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishYtDlp(2, "2026.02.02", epoch.Add(24*time.Hour), "yt-dlp build 2")
	w.fetchErr[w.releases["yt-dlp"].URL] = errors.New("connection reset")
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 0 {
		t.Errorf("updated = %v", updated)
	}
	if _, ok := rec.find(EventUpdateFailed, "yt-dlp"); !ok {
		t.Errorf("events = %s", rec.kinds())
	}
	if read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 1" {
		t.Error("the working copy was changed by a failed download")
	}
	if loadManifest(dir)["yt-dlp"].AssetID != 1 {
		t.Error("the record moved on although nothing was installed")
	}
	if got := listing(t, dir); strings.Contains(got, ".download") || strings.Contains(got, ".part") {
		t.Errorf("leftovers: %s", got)
	}
}

func TestAnUpdateThatFailsItsChecksumKeepsTheWorkingTool(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishYtDlp(2, "2026.02.02", epoch.Add(24*time.Hour), "yt-dlp build 2")
	tampered := w.releases["yt-dlp"]
	tampered.SHA256 = strings.Repeat("f", 64)
	w.releases["yt-dlp"] = tampered
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	failed, ok := rec.find(EventUpdateFailed, "yt-dlp")
	if len(updated) != 0 || !ok || !strings.Contains(failed.Err.Error(), "checksum") {
		t.Fatalf("updated=%v events=%s", updated, rec.kinds())
	}
	if read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 1" {
		t.Error("a build that failed its checksum replaced the working one")
	}
}

func TestAToolThatIsMissingIsInstalledByAnUpdateCheckAndNotCountedAsUpdated(t *testing.T) {
	dir, w := installedWorld(t)
	if err := os.Remove(binPath(dir, "yt-dlp")); err != nil {
		t.Fatal(err)
	}
	i, rec := newTestInstaller(t, dir, w, nil)

	updated := i.update()

	if len(updated) != 0 {
		t.Errorf("updated = %v", updated)
	}
	if _, ok := rec.find(EventInstalled, "yt-dlp"); !ok || read(t, binPath(dir, "yt-dlp")) != "yt-dlp build 1" {
		t.Errorf("events = %s", rec.kinds())
	}
}

func TestAnUpdateWorksWhenTheFolderIsRelativeAndTheProgramIsAlreadyOnPath(t *testing.T) {
	dir, w := installedWorld(t)
	w.publishYtDlp(2, "2026.02.02", epoch.Add(24*time.Hour), "yt-dlp build 2")
	i, _ := newTestInstaller(t, dir, w, nil)
	i.lookPath = exec.LookPath

	if updated := i.update(); len(updated) != 1 {
		t.Fatalf("updated = %v", updated)
	}
	found, err := exec.LookPath("yt-dlp")
	if err != nil || filepath.Dir(found) != dir {
		t.Errorf("PATH finds %q (%v), want the app's folder %q", found, err, dir)
	}
	_ = runtime.GOOS
}
