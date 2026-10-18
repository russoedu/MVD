package engine

import (
	"context"
	"testing"
	"time"
)

func TestHeldSongsDownloadOnlyWhenReleased(t *testing.T) {
	eng := stubEngine(t, true, "https://youtube.com/playlist?list=A")
	eng.opts.HoldForReview = true
	eng.live = liveOf(eng.opts)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	var seen []interface{}
	idle := false
	deadline := time.After(20 * time.Second)
	// Nothing downloads while the songs are held.
	for !idle {
		select {
		case ev := <-eng.Events():
			seen = append(seen, ev)
			switch e := ev.(type) {
			case EvEntryState:
				if e.State == StateDownloading || e.State == StateDone {
					t.Fatalf("a held song reached %v", e.State)
				}
			case EvIdle:
				idle = true
			}
		case <-deadline:
			t.Fatal("the engine never went quiet")
		}
	}

	plan := eng.Plan()
	if len(plan) == 0 {
		t.Fatal("nothing was held")
	}

	// Release the first as proposed; the others stay held.
	if !eng.Release(0, "") {
		t.Fatal("the first song was not released")
	}
	if eng.Release(0, "") {
		t.Error("a song that is already released must not be released twice")
	}
	var first EvEntryState
	for first.State != StateDone {
		select {
		case ev := <-eng.Events():
			if e, ok := ev.(EvEntryState); ok {
				if e.Entry != 0 && (e.State == StateDownloading || e.State == StateDone) {
					t.Fatalf("song %d downloaded without being released", e.Entry)
				}
				if e.Entry == 0 {
					first = e
				}
			}
		case <-deadline:
			t.Fatalf("the released song never finished: %+v", first)
		}
	}

	cancel()
	for range eng.Events() {
	}
	<-done
}
