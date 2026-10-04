package session

import (
	"context"
	"sync/atomic"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func listedThree() engine.EvPlaylistListed {
	return engine.EvPlaylistListed{Playlist: 0, Title: "Mix", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "a", Title: "a"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "b", Title: "b"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "c", Title: "c"},
	}}
}

func TestTheFailureHookRunsOncePerFailedEntryAndOnlyForFailures(t *testing.T) {
	var calls atomic.Int32
	var s *Session
	fakes := &engines{}
	// Reading the snapshot inside the hook would deadlock if a lock were still held.
	s = New(context.Background(), fakes.factory, WithFailureHook(func() {
		_ = s.Snapshot()
		calls.Add(1)
	}))
	t.Cleanup(s.Close)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)

	fake.emit(listedThree())
	fake.emit(engine.EvEntryState{Entry: 0, State: engine.StateDownloading})
	fake.emit(engine.EvEntryState{Entry: 2, State: engine.StateDone})
	fake.emit(engine.EvEntryState{Entry: 0, State: engine.StateFailed, Err: "boom"})
	fake.emit(engine.EvEntryState{Entry: 1, State: engine.StateFailed, Err: "boom"})

	waitFor(t, "both failures to be seen", func() bool { return calls.Load() == 2 })
	waitFor(t, "the last event to be applied", func() bool { return s.Snapshot().Tally.Failed == 2 })
	if calls.Load() != 2 {
		t.Errorf("hook ran %d times, want 2", calls.Load())
	}
}

func TestALinkThatCannotBeLookedUpAlsoCountsAsAFailure(t *testing.T) {
	var calls atomic.Int32
	fakes := &engines{}
	s := New(context.Background(), fakes.factory, WithFailureHook(func() { calls.Add(1) }))
	t.Cleanup(s.Close)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}

	fakes.engine(t).emit(engine.EvPlaylistFailed{Playlist: 0, Err: "This video is unavailable"})

	waitFor(t, "the hook", func() bool { return calls.Load() == 1 })
}

func TestASessionWithoutAHookIsUnaffected(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)

	fake.emit(listedThree())
	fake.emit(engine.EvEntryState{Entry: 0, State: engine.StateFailed, Err: "boom"})

	waitFor(t, "the failure", func() bool { return s.Snapshot().Tally.Failed == 1 })
}

func TestRetryFailedQueuesEveryFailedEntryAgainAndNothingElse(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)
	fake.emit(listedThree())
	fake.emit(engine.EvEntryState{Entry: 0, State: engine.StateFailed, Err: "boom"})
	fake.emit(engine.EvEntryState{Entry: 1, State: engine.StateFailed, Err: "boom"})
	fake.emit(engine.EvEntryState{Entry: 2, State: engine.StateDone})
	waitFor(t, "the failures", func() bool { return s.Snapshot().Tally.Failed == 2 })

	retried := s.RetryFailed()

	fake.mu.Lock()
	asked := append([]int(nil), fake.retried...)
	fake.mu.Unlock()
	if len(asked) != 2 || asked[0] != 0 || asked[1] != 1 {
		t.Errorf("asked to retry %v, want [0 1]", asked)
	}
	// The fake accepts only entry 1, so one of the two was queued again.
	if retried != 1 {
		t.Errorf("retried = %d, want 1", retried)
	}
}

func TestRetryFailedAlsoTriesAgainALinkThatCouldNotBeLookedUp(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/broken", "https://a.example/fine"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)
	fake.emit(engine.EvPlaylistFailed{Playlist: 0, Err: "yt-dlp is out of date"})
	fake.emit(engine.EvPlaylistListed{Playlist: 1, Title: "Fine", Entries: []engine.EntryInfo{{ID: 0, Playlist: 1, Index: 1, VideoID: "a", Title: "a"}}})
	waitFor(t, "the failed listing", func() bool { return s.Snapshot().Playlists[0].Err != "" })

	retried := s.RetryFailed()

	if retried != 1 {
		t.Errorf("retried = %d, want 1 (only the link that failed)", retried)
	}
	broken := 0
	for _, url := range fake.sourceURLs() {
		if url == "https://a.example/broken" {
			broken++
		}
	}
	if broken != 2 {
		t.Errorf("the broken link is queued %d times, want 2 (the first try and the retry): %v", broken, fake.sourceURLs())
	}
}

func TestRetryFailedLeavesALinkAloneOnceItHasWorked(t *testing.T) {
	s, fakes := newSession(t)
	if _, err := s.Add([]string{"https://a.example/1"}); err != nil {
		t.Fatal(err)
	}
	fake := fakes.engine(t)
	fake.emit(engine.EvPlaylistFailed{Playlist: 0, Err: "boom"})
	waitFor(t, "the failure", func() bool { return s.Snapshot().Playlists[0].Err != "" })
	if first := s.RetryFailed(); first != 1 {
		t.Fatalf("first retry = %d", first)
	}
	fake.emit(engine.EvPlaylistListed{Playlist: 1, Title: "Now fine"})
	waitFor(t, "the second try to be listed", func() bool { return len(s.Snapshot().Playlists) == 2 && s.Snapshot().Playlists[1].Title == "Now fine" })

	if again := s.RetryFailed(); again != 0 {
		t.Errorf("a link that has since been listed was queued again: %d", again)
	}
}

func TestRetryFailedBeforeAnyEngineIsZero(t *testing.T) {
	s, _ := newSession(t)

	if got := s.RetryFailed(); got != 0 {
		t.Errorf("retried = %d", got)
	}
}
