package toolupdates

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"youtube-downloader/apps/mvd/notification"
)

// failureCheckGap is the shortest time between two checks that a failed download
// starts. A playlist can fail many entries in a row, and GitHub limits how often it
// can be asked, so one check covers them all.
const failureCheckGap = 10 * time.Minute

// Updates decides when to look for newer builds of yt-dlp and ffmpeg: once at
// start-up, and again when a download fails, since an out-of-date yt-dlp is the usual
// reason a download that worked yesterday does not today.
//
// Checks run in the background and never hold up the page or a download. At most one
// runs at a time.
type Updates struct {
	// run checks for and installs newer builds, and returns the names it replaced.
	run func() []string
	// retry queues every failed download again and says how many it queued.
	retry  func() int
	notify func(notification.Notice)
	logf   func(string, ...interface{})
	now    func() time.Time

	mu            sync.Mutex
	running       bool
	lastFromError time.Time
	// done, when set, is called after each check finishes. Tests use it to wait.
	done func()
}

func New(run func() []string, retry func() int, notify func(notification.Notice), logf func(string, ...interface{}), now func() time.Time) *Updates {
	return &Updates{run: run, retry: retry, notify: notify, logf: logf, now: now}
}

// AtStart looks for updates in the background.
func (t *Updates) AtStart() {
	if t.begin(false) {
		go t.check(false)
	}
}

// AfterFailure looks for updates in the background, unless a check is already running or
// one that a failure started finished less than failureCheckGap ago. It is cheap and
// returns at once, because the session calls it from the goroutine that keeps the page
// up to date.
func (t *Updates) AfterFailure() {
	if t.begin(true) {
		go t.check(true)
	}
}

// begin claims the right to run a check, and says whether it was granted.
func (t *Updates) begin(fromFailure bool) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return false
	}
	if fromFailure && !t.lastFromError.IsZero() && t.now().Sub(t.lastFromError) < failureCheckGap {
		return false
	}
	t.running = true
	if fromFailure {
		t.lastFromError = t.now()
	}

	return true
}

// check runs one check. When it replaced something after a failure, the failed
// downloads are queued again, since they may well work now.
func (t *Updates) check(fromFailure bool) {
	defer func() {
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
		if t.done != nil {
			t.done()
		}
	}()

	updated := t.run()
	if len(updated) == 0 {
		return
	}
	list := strings.Join(updated, " and ")
	t.logf("Updated %s", list)

	text := "Downloads you start from now on use the new version."
	if fromFailure {
		retried := t.retry()
		t.logf("Trying %d failed download(s) again with the new version", retried)
		if retried > 0 {
			text = fmt.Sprintf("A download had failed, so %d failed download(s) are being tried again with the new version.", retried)
		} else {
			text = "A download had failed. Add it again to try it with the new version."
		}
	}
	t.notify(notification.Notice{Title: "MVD updated " + list, Text: text})
}
