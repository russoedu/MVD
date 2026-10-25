package engine

import (
	"context"
	"testing"
	"time"
)

// untilNext reads events until one satisfies done, or fails after a while.
func untilNext(t *testing.T, eng *Engine, done func(ev interface{}) bool) {
	t.Helper()
	timeout := time.After(20 * time.Second)
	for {
		select {
		case ev, ok := <-eng.Events():
			if !ok {
				t.Fatal("events closed")
			}
			if done(ev) {
				return
			}
		case <-timeout:
			t.Fatal("timed out")
		}
	}
}

func TestASavedSongCanBeSwappedForAnotherVideo(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=A")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go eng.Run(ctx)

	untilNext(t, eng, func(ev interface{}) bool { _, idle := ev.(EvIdle); return idle })
	if eng.Swap(0, "") {
		t.Error("a swap with no video must not be taken")
	}
	if !eng.Swap(0, "swapped12345") {
		t.Fatal("a saved song could not be swapped")
	}
	var chosen string
	untilNext(t, eng, func(ev interface{}) bool {
		if e, ok := ev.(EvEntryState); ok && e.Entry == 0 && e.State == StateAccepted {
			chosen = e.TargetID
		}
		_, idle := ev.(EvIdle)

		return idle
	})
	if chosen != "swapped12345" {
		t.Errorf("the song was accepted with %q", chosen)
	}
	for _, snap := range eng.Snapshot() {
		if snap.Upload.ID != "" && snap.TargetID == "swapped12345" && snap.State == StateDone && snap.Kind == KindChosen {
			return
		}
	}
	t.Error("the swapped song did not end up saved as the person's choice")
}

func TestARemovedLinkIsDroppedAndLeftOutOfTheSnapshot(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=A")
	eng.opts.HoldForReview = true
	eng.live = liveOf(eng.opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go eng.Run(ctx)

	untilNext(t, eng, func(ev interface{}) bool { _, idle := ev.(EvIdle); return idle })
	if !eng.RemoveSource(0) {
		t.Fatal("the link could not be removed")
	}
	if eng.RemoveSource(0) {
		t.Error("a link was removed twice")
	}
	if eng.Release(0, "") || eng.Swap(0, "abc12345678") {
		t.Error("a song of a removed link must not be sent to download")
	}
	if got := len(eng.Snapshot()); got != 0 {
		t.Errorf("the snapshot still has %d songs of the removed link", got)
	}
	untilNext(t, eng, func(ev interface{}) bool { r, ok := ev.(EvPlaylistRemoved); return ok && r.Playlist == 0 })
}

func TestAnAcceptedSongWaitsForTheDownloadsToStartAndCanBeRejectedBefore(t *testing.T) {
	eng := stubEngine(t, false, "https://youtube.com/playlist?list=A")
	eng.opts.HoldForReview = true
	eng.live = liveOf(eng.opts)
	eng.Pause()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go eng.Run(ctx)
	untilNext(t, eng, func(ev interface{}) bool { _, idle := ev.(EvIdle); return idle })

	if !eng.Release(0, "") || !eng.Release(1, "") {
		t.Fatal("two songs could not be accepted")
	}
	if !eng.Skip(1) {
		t.Error("an accepted song could not be rejected before its turn")
	}
	time.Sleep(300 * time.Millisecond)
	for _, snap := range eng.Snapshot() {
		if snap.State == StateDownloading || snap.State == StateDone {
			t.Fatalf("a song went on while the downloads were stopped: %v", snap.State)
		}
	}
	eng.Resume()
	untilNext(t, eng, func(ev interface{}) bool {
		e, ok := ev.(EvEntryState)

		return ok && e.Entry == 0 && e.State == StateDone
	})
}
