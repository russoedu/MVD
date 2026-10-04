package session

import (
	"context"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestSnapshotBeforeAnyEngineIsEmptyAndIdle(t *testing.T) {
	s, fakes := newSession(t)

	snap := s.Snapshot()

	if len(snap.Playlists) != 0 || len(snap.Entries) != 0 || !snap.Idle {
		t.Errorf("want an empty idle snapshot, got %+v", snap)
	}
	if len(fakes.calls) != 0 {
		t.Error("an engine was built before any URL arrived")
	}
}

func TestVersionAdvancesWithEveryEventAndNeverGoesBack(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)
	before := s.Snapshot().Version

	fake.emit(engine.EvPlaylistListing{Playlist: 0, URL: "https://a.example/1"})
	fake.emit(engine.EvLog{Playlist: 0, Entry: -1, Line: "listing"})

	waitFor(t, "two more versions", func() bool { return s.Snapshot().Version >= before+2 })
}

func TestEventsReachTheSnapshot(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	fakes.engine(t).emit(engine.EvPlaylistListed{Playlist: 0, Title: "My Mix", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, VideoID: "v0", Title: "zero"}}})

	waitFor(t, "the listing", func() bool {
		snap := s.Snapshot()
		return len(snap.Playlists) == 1 && snap.Playlists[0].Title == "My Mix" && len(snap.Entries) == 1
	})
}

func TestWaitSinceReturnsAtOnceWhenTheVersionIsAlreadyNewer(t *testing.T) {
	s, _ := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	snap, ok := s.WaitSince(context.Background(), 0)

	if !ok || snap.Version <= 0 {
		t.Errorf("WaitSince(0) = %+v, %v; want a newer snapshot at once", snap, ok)
	}
}

func TestWaitSinceBlocksUntilSomethingChanges(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	current := s.Snapshot().Version
	got := make(chan int64, 1)
	go func() {
		snap, ok := s.WaitSince(context.Background(), current)
		if ok {
			got <- snap.Version
		}
	}()

	select {
	case v := <-got:
		t.Fatalf("WaitSince returned %d before anything changed", v)
	case <-time.After(50 * time.Millisecond):
	}

	fakes.engine(t).emit(engine.EvLog{Playlist: 0, Entry: -1, Line: "x"})

	select {
	case v := <-got:
		if v <= current {
			t.Errorf("version %d is not newer than %d", v, current)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("WaitSince never returned after a change")
	}
}

func TestWaitSinceGivesUpWhenItsContextEnds(t *testing.T) {
	s, _ := newSession(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	if _, ok := s.WaitSince(ctx, s.Snapshot().Version); ok {
		t.Error("WaitSince reported a change that never happened")
	}
}

func TestCloseStopsTheEngineAndKeepsEveryEventItHadSent(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fakes.engine(t).emit(engine.EvPlaylistFailed{Playlist: 0, Err: "gone"})

	s.Close()

	if got := s.Snapshot().Playlists[0].Err; got != "gone" {
		t.Errorf("an event sent before Close was lost: err = %q", got)
	}
}

func TestAddRefusesAfterCloseAndCloseIsRepeatable(t *testing.T) {
	s, _ := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	s.Close()
	s.Close()

	if _, err := s.Add([]string{"https://b.example/2"}); err != ErrClosed {
		t.Errorf("Add after Close = %v, want ErrClosed", err)
	}
}

func TestCloseBeforeAnyEngineIsFine(t *testing.T) {
	s, fakes := newSession(t)

	s.Close()

	if _, err := s.Add([]string{"https://a.example/1"}); err != ErrClosed {
		t.Errorf("Add after Close = %v, want ErrClosed", err)
	}
	if len(fakes.calls) != 0 {
		t.Error("an engine was built after Close")
	}
}
