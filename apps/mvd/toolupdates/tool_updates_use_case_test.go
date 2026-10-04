package toolupdates

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/libs/mvd-core/deps"
)

// updatesProbe is an Updates over a fake checker, with a clock the test moves.
type updatesProbe struct {
	updates *Updates
	runs    atomic.Int32
	retries atomic.Int32

	mu      sync.Mutex
	notices []notification.Notice
	now     time.Time

	// result is what each check returns.
	result []string
	// release, when set, holds a check open until it is closed.
	release chan struct{}
	done    chan struct{}
}

func newUpdatesProbe(result ...string) *updatesProbe {
	p := &updatesProbe{result: result, now: time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC), done: make(chan struct{}, 8)}
	p.updates = New(
		func() []string {
			p.runs.Add(1)
			if p.release != nil {
				<-p.release
			}

			return p.result
		},
		func() int { p.retries.Add(1); return 3 },
		func(n notification.Notice) { p.mu.Lock(); p.notices = append(p.notices, n); p.mu.Unlock() },
		func(string, ...interface{}) {},
		func() time.Time { p.mu.Lock(); defer p.mu.Unlock(); return p.now },
	)
	p.updates.done = func() { p.done <- struct{}{} }

	return p
}

func (p *updatesProbe) advance(d time.Duration) { p.mu.Lock(); p.now = p.now.Add(d); p.mu.Unlock() }

func (p *updatesProbe) wait(t *testing.T) {
	t.Helper()
	select {
	case <-p.done:
	case <-time.After(2 * time.Second):
		t.Fatal("the check did not finish")
	}
}

func (p *updatesProbe) noticeList() []notification.Notice {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]notification.Notice(nil), p.notices...)
}

func TestACheckAtStartRunsAndTellsThePersonWhatWasUpdatedWithoutRetryingAnything(t *testing.T) {
	p := newUpdatesProbe("yt-dlp", "ffmpeg")

	p.updates.AtStart()
	p.wait(t)

	if p.runs.Load() != 1 || p.retries.Load() != 0 {
		t.Errorf("runs=%d retries=%d", p.runs.Load(), p.retries.Load())
	}
	notices := p.noticeList()
	if len(notices) != 1 || notices[0].Title != "MVD updated yt-dlp and ffmpeg" || notices[0].Failure {
		t.Fatalf("notices = %+v", notices)
	}
}

func TestACheckThatFindsNothingNewSaysNothingAndRetriesNothing(t *testing.T) {
	p := newUpdatesProbe()

	p.updates.AtStart()
	p.wait(t)
	p.updates.AfterFailure()
	p.wait(t)

	if len(p.noticeList()) != 0 || p.retries.Load() != 0 {
		t.Errorf("notices=%+v retries=%d", p.noticeList(), p.retries.Load())
	}
}

func TestAfterAFailureAnUpdateQueuesTheFailedDownloadsAgainAndSaysHowMany(t *testing.T) {
	p := newUpdatesProbe("yt-dlp")

	p.updates.AfterFailure()
	p.wait(t)

	if p.retries.Load() != 1 {
		t.Errorf("retries = %d, want 1", p.retries.Load())
	}
	notices := p.noticeList()
	if len(notices) != 1 || !strings.Contains(notices[0].Text, "3 failed download") {
		t.Errorf("notices = %+v", notices)
	}
}

func TestFailuresThatPileUpWhileACheckIsRunningStartOnlyOneCheck(t *testing.T) {
	p := newUpdatesProbe()
	p.release = make(chan struct{})

	for range 50 {
		p.updates.AfterFailure()
	}
	close(p.release)
	p.wait(t)

	if p.runs.Load() != 1 {
		t.Errorf("%d checks ran for 50 failures, want 1", p.runs.Load())
	}
}

func TestAFailureRightAfterACheckDoesNotStartAnotherUntilTheGapHasPassed(t *testing.T) {
	p := newUpdatesProbe()

	p.updates.AfterFailure()
	p.wait(t)

	p.advance(failureCheckGap - time.Second)
	if p.updates.begin(true) {
		t.Fatalf("a second check was allowed %v after the first, inside the gap", failureCheckGap-time.Second)
	}

	p.advance(2 * time.Second)
	if !p.updates.begin(true) {
		t.Fatal("a check was refused after the gap had passed")
	}
	p.updates.check(true)
	p.wait(t)

	if p.runs.Load() != 2 {
		t.Errorf("runs = %d, want 2 (the first, and the one after the gap)", p.runs.Load())
	}
}

func TestOnlyOneCheckRunsAtATimeWhateverStartedIt(t *testing.T) {
	p := newUpdatesProbe()

	if !p.updates.begin(false) {
		t.Fatal("the first check was refused")
	}
	for _, fromFailure := range []bool{false, true} {
		if p.updates.begin(fromFailure) {
			t.Errorf("a check (from a failure: %v) was allowed while another was running", fromFailure)
		}
	}

	p.updates.check(false)
	p.wait(t)
	if !p.updates.begin(false) {
		t.Error("a check was refused after the running one finished")
	}
}

func TestACheckAtStartIsNotLimitedByTheGapThatFollowsAFailureCheck(t *testing.T) {
	p := newUpdatesProbe()
	p.updates.AfterFailure()
	p.wait(t)

	if !p.updates.begin(false) {
		t.Error("a start check was refused because of the gap that only limits failure checks")
	}
}

func TestCheckingForUpdatesOnlyLogsAndNeverNotifiesByItself(t *testing.T) {
	var notices []notification.Notice
	var logs []string
	report := Reporter(
		func(format string, a ...interface{}) { logs = append(logs, format) },
		func(n notification.Notice) { notices = append(notices, n) },
	)

	report(deps.Event{Kind: deps.EventUpToDate, Name: "yt-dlp", Detail: "2026.08.19"})
	report(deps.Event{Kind: deps.EventUpdated, Name: "yt-dlp", Detail: "2026.09.30"})
	report(deps.Event{Kind: deps.EventUpdateFailed, Name: "ffmpeg", Err: errors.New("GitHub is limiting requests")})

	if len(logs) != 3 || len(notices) != 0 {
		t.Errorf("logs=%d notices=%+v", len(logs), notices)
	}
}
