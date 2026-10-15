package engine

import (
	"context"
	"testing"
	"time"
)

func TestThrottleLosesADownloaderEveryFiveRateLimits(t *testing.T) {
	th := newDownloadThrottle(3)
	for i := 0; i < 4; i++ {
		if msg := th.rateLimited(); msg != "" {
			t.Fatalf("failure %d changed things: %s", i+1, msg)
		}
	}
	if msg := th.rateLimited(); msg == "" || th.limit != 2 {
		t.Fatalf("fifth failure: limit %d, %q", th.limit, msg)
	}
	for i := 0; i < 5; i++ {
		th.rateLimited()
	}
	if th.limit != 1 || th.wait != 0 {
		t.Fatalf("limit %d wait %s, want 1 and no wait", th.limit, th.wait)
	}
}

func TestThrottleWaitGrowsAtOneDownloaderAndRecovers(t *testing.T) {
	th := newDownloadThrottle(2)
	for i := 0; i < 5; i++ {
		th.rateLimited()
	}
	th.rateLimited()
	first := th.wait
	th.rateLimited()
	if first != throttleFirstWait || th.wait <= first {
		t.Fatalf("wait %s then %s: it should start at %s and grow", first, th.wait, throttleFirstWait)
	}
	for i := 0; i < 60 && (th.wait > 0 || th.limit < 2); i++ {
		th.succeeded()
	}
	if th.wait != 0 || th.limit != 2 {
		t.Fatalf("not recovered: limit %d wait %s", th.limit, th.wait)
	}
}

func TestThrottleHoldsBackAboveTheLimit(t *testing.T) {
	th := newDownloadThrottle(2)
	for i := 0; i < 5; i++ {
		th.rateLimited()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	release, ok := th.acquire(ctx)
	if !ok {
		t.Fatal("the first downloader should be given")
	}
	if _, ok := th.acquire(ctx); ok {
		t.Fatal("a second downloader was given with the limit at 1")
	}
	release()
}

func TestIsRateLimited(t *testing.T) {
	for msg, want := range map[string]bool{
		"ERROR: HTTP Error 429: Too Many Requests": true,
		"unable to download: HTTP Error 409":       true,
		"HTTP Error 403: Forbidden":                false,
		"Video unavailable":                        false,
	} {
		if got := isRateLimited(msg); got != want {
			t.Errorf("isRateLimited(%q) = %v", msg, got)
		}
	}
}
