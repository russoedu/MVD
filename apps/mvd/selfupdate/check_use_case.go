package selfupdate

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/question"
)

const (
	answerUpdate = "Update now"
	answerLater  = "Not now"
)

// Checker looks for a newer MVD: by itself at start, at most once a day and only when the
// person allows it, and whenever they ask from the Help menu.
type Checker struct {
	// Dir is the app-data folder, where the last look is remembered.
	Dir string
	// Version is the running version.
	Version string
	// Enabled says whether the person allows the automatic look; read each time.
	Enabled func() bool
	// Notify tells the person something without taking over their screen.
	Notify func(notification.Notice)
	// Ask puts a question with these answers to the person, in the system's own dialog.
	Ask func(title, text string, choices []string) question.Answer
	// Open shows an address in the person's browser.
	Open func(url string) error
	// Logf writes to the log.
	Logf func(format string, args ...interface{})
	// Quit asks the program to close, as the close button does: it asks first when songs download.
	// The new version starts once the program has closed.
	Quit func()
	// Client, URL and Now are for the tests; the zero values are the internet, GitHub's list and
	// the clock.
	Client *http.Client
	URL    string
	Now    func() time.Time

	latest atomic.Value // string: the newest version known so far
	// relaunch is the program to start when this one has closed, once an update was put in place.
	relaunch atomic.Value // string

	// The following are for the tests.
	executable func() (string, error)
	goos       string
	goarch     string
	insecure   bool // allow http addresses
}

// Relaunch is the program to start now that the app has closed, or "" when no update is waiting.
func (c *Checker) Relaunch() string {
	v, _ := c.relaunch.Load().(string)

	return v
}

func (c *Checker) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}

	return time.Now()
}

func (c *Checker) fetch(ctx context.Context) (Release, error) {
	client, url := c.Client, c.URL
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if url == "" {
		url = releasesURL
	}

	return newestRelease(ctx, client, url)
}

// Note is the line the About window adds when a newer version is known, or "". It reads what the
// last look found, so it is quick and never goes to the network.
func (c *Checker) Note() string {
	if c.latest.Load() == nil {
		c.latest.Store(loadState(c.Dir).Latest)
	}
	if v, _ := c.latest.Load().(string); isNewer(c.Version, v) {
		return fmt.Sprintf("MVD %s is available: Help > Check for updates.", v)
	}

	return ""
}

// AtStart looks once a day, when the person allows it, and tells them once about each newer
// version. It runs on the caller's goroutine and is silent when there is no connection.
func (c *Checker) AtStart(ctx context.Context) {
	if c.Enabled != nil && !c.Enabled() {
		return
	}
	state := loadState(c.Dir)
	if !due(state.CheckedAt, c.now()) {
		return
	}
	release, err := c.fetch(ctx)
	if err != nil {
		c.Logf("Could not look for a newer MVD: %v", err)

		return
	}
	state.CheckedAt, state.Latest = c.now(), release.Version
	c.latest.Store(release.Version)
	if isNewer(c.Version, release.Version) && state.Told != release.Version {
		state.Told = release.Version
		c.Notify(notification.Notice{
			Title: "MVD " + release.Version + " is available",
			Text:  fmt.Sprintf("You have %s. Choose Help > Check for updates in MVD to get it.", c.Version),
		})
	}
	saveState(c.Dir, state)
}

// CheckNow looks right away, whatever the day, and asks about a newer version. It returns the
// line to show on the screen afterwards, or "" when there is nothing to add.
func (c *Checker) CheckNow(ctx context.Context) string {
	if _, ok := parseVersion(c.Version); !ok {
		return "This build has no version number, so it cannot look for updates."
	}
	release, err := c.fetch(ctx)
	if err != nil {
		c.Logf("Could not look for a newer MVD: %v", err)

		return "Could not look for updates: " + err.Error()
	}
	state := loadState(c.Dir)
	state.CheckedAt, state.Latest = c.now(), release.Version
	c.latest.Store(release.Version)
	defer func() { saveState(c.Dir, state) }()

	if !isNewer(c.Version, release.Version) {
		return fmt.Sprintf("MVD %s is up to date.", c.Version)
	}
	state.Told = release.Version
	text := fmt.Sprintf("MVD %s is available. You have %s.\n\nUpdate now?", release.Version, c.Version)
	switch c.Ask("MVD", text, []string{answerUpdate, answerLater}) {
	case question.AnswerFirst:
		return c.update(ctx, release)
	case question.AnswerUnavailable:
		// Nothing could be shown, so take the person to where the new version is.
		return c.showPage(release)
	}

	return fmt.Sprintf("MVD %s is available. Not updating now.", release.Version)
}

// update brings the release in: the zip for this system is downloaded, checked against the
// checksum GitHub states for it, and the running program is replaced by the one inside. When
// that cannot be done safely (another system, a folder that is not writable, a download that
// fails or does not match) nothing is changed and the person is taken to the release page.
func (c *Checker) update(ctx context.Context, release Release) string {
	if err := c.replaceFrom(ctx, release); err != nil {
		c.Logf("Could not update MVD by itself: %v", err)

		return fmt.Sprintf("MVD could not update itself (%v). %s", err, c.showPage(release))
	}

	return fmt.Sprintf("MVD %s is in place. It opens when you close this one.", release.Version)
}

func (c *Checker) replaceFrom(ctx context.Context, release Release) error {
	goos, goarch := c.goos, c.goarch
	if goos == "" {
		goos, goarch = runtime.GOOS, runtime.GOARCH
	}
	name, ok := assetName(release.Version, goos, goarch)
	if !ok {
		return fmt.Errorf("it only does that on Windows so far, and this is %s", goos)
	}
	asset, ok := assetOf(release, name)
	if !ok {
		return fmt.Errorf("%s is not in the release", name)
	}
	if !c.insecure && !strings.HasPrefix(asset.URL, "https://") {
		return fmt.Errorf("the download is not an https address")
	}
	executable := c.executable
	if executable == nil {
		executable = os.Executable
	}
	exe, err := executable()
	if err != nil {
		return err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return err
	}

	// The pieces are kept in the app-data folder, which is always writable, and removed after.
	work, err := os.MkdirTemp(c.Dir, "update-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(work) }()
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	zipPath, newProgram := filepath.Join(work, name), filepath.Join(work, programName)
	if err := downloadChecked(ctx, client, asset.URL, asset.Digest, zipPath); err != nil {
		return err
	}
	if err := extractProgram(zipPath, programName, newProgram); err != nil {
		return err
	}
	if err := replaceProgram(exe, newProgram); err != nil {
		return err
	}
	c.relaunch.Store(exe)
	if c.Quit != nil {
		c.Quit()
	}

	return nil
}

func (c *Checker) showPage(release Release) string {
	if err := c.Open(release.Page); err != nil {
		return fmt.Sprintf("MVD %s is available: %s", release.Version, release.Page)
	}

	return fmt.Sprintf("Opened the page of MVD %s in your browser.", release.Version)
}
