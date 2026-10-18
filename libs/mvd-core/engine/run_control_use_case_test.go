package engine

import (
	"context"
	"testing"
	"time"
)

func TestRaisingTheMaximumTakesEffectAtOnce(t *testing.T) {
	th := newDownloadThrottle(2)
	th.setMax(5)
	if th.limit != 5 || th.max != 5 {
		t.Fatalf("limit %d max %d, want 5 and 5", th.limit, th.max)
	}
}

func TestLoweringTheMaximumLetsTheRunningOnesFinish(t *testing.T) {
	th := newDownloadThrottle(4)
	ctx := context.Background()
	var release []func()
	for i := 0; i < 4; i++ {
		r, ok := th.acquire(ctx)
		if !ok {
			t.Fatal("no downloader")
		}
		release = append(release, r)
	}
	th.setMax(2)

	// Nothing is interrupted: four are still running, and no new one may start until
	// they have dropped under the new number.
	started := make(chan struct{})
	go func() {
		if r, ok := th.acquire(ctx); ok {
			close(started)
			r()
		}
	}()
	release[0]()
	release[1]()
	select {
	case <-started:
		t.Fatal("a download started with two still running and a maximum of two")
	case <-time.After(100 * time.Millisecond):
	}
	release[2]()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("a download should start once fewer than two are running")
	}
	release[3]()
}

func TestAThrottleHeldBackByTheServiceKeepsItsLimitWhenTheMaximumRises(t *testing.T) {
	th := newDownloadThrottle(4)
	for i := 0; i < 5; i++ {
		th.rateLimited()
	}
	if th.limit != 3 {
		t.Fatalf("limit %d, want 3 after five rate limits", th.limit)
	}
	th.setMax(8)
	if th.limit != 3 || th.max != 8 {
		t.Errorf("limit %d max %d: the service's hold must stay, with room to grow back", th.limit, th.max)
	}
	th.setMax(2)
	if th.limit != 2 {
		t.Errorf("limit %d, want 2: never above the maximum", th.limit)
	}
}

func TestPausedStopsNewDownloadsAndResumeLetsThemGo(t *testing.T) {
	th := newDownloadThrottle(2)
	th.setPaused(true)
	got := make(chan struct{})
	go func() {
		if r, ok := th.acquire(context.Background()); ok {
			close(got)
			r()
		}
	}()
	select {
	case <-got:
		t.Fatal("a download started while paused")
	case <-time.After(100 * time.Millisecond):
	}
	th.setPaused(false)
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("the download did not start after resuming")
	}
}

func TestReconfigureChangesTheLiveSettings(t *testing.T) {
	eng := stubEngine(t, false)
	s := eng.Settings()
	s.Workers, s.Quality, s.RetryCycles = 7, "best[height<=720]", 3
	eng.Reconfigure(s)
	got := eng.Settings()
	if got.Workers != 7 || got.Quality != "best[height<=720]" || got.RetryCycles != 3 {
		t.Errorf("settings %+v", got)
	}
	if eng.throttle.max != 7 {
		t.Errorf("throttle max %d, want 7", eng.throttle.max)
	}
	s.Workers = 1000
	eng.Reconfigure(s)
	if eng.throttle.max != maxDownloaders {
		t.Errorf("throttle max %d, want the ceiling %d", eng.throttle.max, maxDownloaders)
	}
}
