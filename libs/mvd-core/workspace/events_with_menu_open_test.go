package workspace

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
)

func TestTheRunsEventsKeepComingWhileAMenuIsOpen(t *testing.T) {
	m, fake := started(t)
	_, wait := m.screen.Attach(fake) // the command the screen waits for an event with
	go func() { fake.events <- engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: 0, URL: "https://youtu.be/x"}} }()
	event := wait()

	m.bar.SetOpen(0)
	next, cmd := m.Update(event)
	if cmd == nil {
		t.Fatal("the screen was not asked for the next event: the run's events pile up until it stops")
	}
	if got := len(next.(Model).screen.State().Playlists); got != 1 {
		t.Errorf("the event was not taken by the screen: %d playlists", got)
	}
}

func TestWhatAnEarlierRunStillSaysIsNotTakenForTheRunNow(t *testing.T) {
	m, fake := started(t)
	old := newFakeRun()
	_, wait := m.screen.Attach(old)
	go func() { old.events <- engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: 5, URL: "https://youtu.be/old"}} }()
	stale := wait()
	_, waitMore := m.screen.Attach(old)
	go old.Close()
	closed := waitMore()

	next, _ := m.Update(stale)
	m = next.(Model)
	next, _ = m.Update(closed)
	m = next.(Model)
	if !m.running() || !m.screen.Running() {
		t.Error("the closing of an earlier run ended the run that is shown")
	}
	if got := len(m.screen.State().Playlists); got != 0 {
		t.Errorf("an event of an earlier run reached the run that is shown: %d playlists", got)
	}
	_ = fake
}
