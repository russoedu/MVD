package songid

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentRequestsAreSpacedByTheInterval(t *testing.T) {
	var p pacer
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.wait(40 * time.Millisecond)
		}()
	}
	wg.Wait()

	// Five slots 40 ms apart: the last starts at 160 ms.
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Errorf("five requests were done in %s, they should be spaced", elapsed)
	}
}

func TestABlockedServiceStaysBlockedUntilItsTimeIsUp(t *testing.T) {
	var p pacer
	if p.blocked() {
		t.Error("a new pacer blocks nothing")
	}
	p.block(40 * time.Millisecond)
	if !p.blocked() {
		t.Error("the service should be left alone")
	}
	time.Sleep(60 * time.Millisecond)
	if p.blocked() {
		t.Error("the block should be over")
	}
}

func TestRefusals(t *testing.T) {
	for _, status := range []int{403, 429, 503} {
		if !refusedBy(status) {
			t.Errorf("%d is a refusal", status)
		}
	}
	if refusedBy(200) || refusedBy(404) {
		t.Error("200 and 404 are not refusals")
	}
}
