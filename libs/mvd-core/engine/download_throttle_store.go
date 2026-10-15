package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// throttleStep is how many rate-limit failures take one downloader away, and how many
	// successes give one back (or shorten the wait).
	throttleStep = 5
	// throttleFirstWait is the wait added the first time a single downloader still fails.
	throttleFirstWait = 2 * time.Second
	// throttleMaxWait caps the wait between downloads.
	throttleMaxWait = 2 * time.Minute
)

// rateLimitSignals mark a failure as the service asking for fewer requests.
var rateLimitSignals = []string{
	"http error 429",
	"http error 409",
	"too many requests",
	"status code 429",
	"status code 409",
}

// isRateLimited reports whether a yt-dlp error is a 429 or 409 answer.
func isRateLimited(errMsg string) bool {
	s := strings.ToLower(errMsg)
	for _, sig := range rateLimitSignals {
		if strings.Contains(s, sig) {
			return true
		}
	}
	return false
}

// downloadThrottle makes the number of parallel downloads a maximum. Every throttleStep
// rate-limit failures take one downloader away, down to one; a single downloader that
// keeps failing gets a wait before each download that grows with every failure. Every
// throttleStep successes undo it: the wait shrinks first, then downloaders come back,
// up to the maximum.
type downloadThrottle struct {
	max       int
	firstWait time.Duration
	maxWait   time.Duration

	mu        sync.Mutex
	limit     int           // downloaders allowed now, 1..max
	wait      time.Duration // wait before each download, 0 when none
	active    int
	failures  int // rate-limit failures since the last step down
	successes int // successes since the last step up or the last failure
	changed   chan struct{}
}

func newDownloadThrottle(max int) *downloadThrottle {
	max = atLeast(max, 1)
	return &downloadThrottle{
		max: max, limit: max,
		firstWait: throttleFirstWait, maxWait: throttleMaxWait,
		changed: make(chan struct{}),
	}
}

// broadcast wakes everything waiting for a slot; t.mu must be held.
func (t *downloadThrottle) broadcast() {
	close(t.changed)
	t.changed = make(chan struct{})
}

// acquire waits for a free downloader and for the wait the throttle asks for. It returns
// the function that gives the downloader back, or false when ctx ended first.
func (t *downloadThrottle) acquire(ctx context.Context) (release func(), ok bool) {
	for {
		t.mu.Lock()
		if t.active < t.limit {
			t.active++
			wait := t.wait
			t.mu.Unlock()
			release = func() {
				t.mu.Lock()
				t.active--
				t.broadcast()
				t.mu.Unlock()
			}
			if wait > 0 && !sleepCtx(ctx, wait) {
				release()
				return nil, false
			}
			return release, true
		}
		changed := t.changed
		t.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return nil, false
		}
	}
}

// succeeded records a finished download and says what changed, "" for nothing.
func (t *downloadThrottle) succeeded() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.limit == t.max && t.wait == 0 {
		t.failures = 0
		return ""
	}
	t.successes++
	if t.successes < throttleStep {
		return ""
	}
	t.successes = 0
	defer t.broadcast()
	if t.wait > 0 {
		t.wait /= 2
		if t.wait < t.firstWait/2 {
			t.wait = 0
		}
		return fmt.Sprintf("downloads are recovering: waiting %s between downloads", t.wait)
	}
	t.limit++
	return fmt.Sprintf("downloads are recovering: %d at a time (up to %d)", t.limit, t.max)
}

// rateLimited records a 429 or 409 and says what changed, "" for nothing.
func (t *downloadThrottle) rateLimited() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.successes = 0
	if t.limit == 1 {
		// Down to one downloader: every failure now lengthens the wait.
		if t.wait == 0 {
			t.wait = t.firstWait
		} else {
			t.wait = min(t.wait+t.wait/2, t.maxWait)
		}
		return fmt.Sprintf("the service is limiting requests: waiting %s between downloads", t.wait)
	}
	t.failures++
	if t.failures < throttleStep {
		return ""
	}
	t.failures = 0
	t.limit--
	t.broadcast()
	return fmt.Sprintf("the service is limiting requests: %d at a time (up to %d)", t.limit, t.max)
}
