package terminalui

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestEventsPassThroughInOrderAndFailuresAreAnnounced(t *testing.T) {
	in := make(chan interface{})
	failures := 0
	out := watchEvents(in, ledgerWithLinks(), func() { failures++ })

	events := []interface{}{
		engine.EvEntryState{Entry: 1, State: engine.StateDownloading},
		engine.EvEntryState{Entry: 1, State: engine.StateFailed},
		engine.EvPlaylistFailed{Playlist: 0, Err: "x"},
	}
	go func() {
		for _, ev := range events {
			in <- ev
		}
		close(in)
	}()

	var got []interface{}
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != len(events) {
		t.Fatalf("got %d events, want %d", len(got), len(events))
	}
	for i := range events {
		if got[i] != events[i] {
			t.Errorf("event %d: got %v, want %v", i, got[i], events[i])
		}
	}
	if failures != 2 {
		t.Errorf("announced %d failures, want 2", failures)
	}
}
