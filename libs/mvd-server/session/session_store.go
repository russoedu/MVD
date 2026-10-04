package session

import (
	"context"
	"errors"
	"sync"

	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-server/snapshot"
)

// ErrClosed is returned once the session has been closed.
var ErrClosed = errors.New("session is closed")

// Session is the one run a front end watches. The zero engine state is "nothing
// yet": it is built when the first URLs arrive and then lives until Close.
type Session struct {
	factory Factory
	root    context.Context

	// adding serialises building the engine against adding to it, so two
	// first pastes cannot build two engines.
	adding sync.Mutex
	closed bool

	mu      sync.Mutex
	state   *runstate.State
	version int64
	changed chan struct{} // closed, and replaced, every time version moves
	eng     Engine
	cancel  context.CancelFunc
	pumped  chan struct{} // closed once every event has been folded into state
	done    chan struct{} // closed once Run has returned
}

// New returns a session whose engine, once built, lives as long as root.
func New(root context.Context, factory Factory) *Session {
	empty := runstate.New(nil)
	// Nothing is running, which is what idle means to a viewer.
	empty.Idle = true
	return &Session{
		factory: factory,
		root:    root,
		state:   empty,
		changed: make(chan struct{}),
	}
}

// Snapshot is the run as it is now.
func (s *Session) Snapshot() snapshot.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return snapshot.From(s.state, s.version)
}

// WaitSince returns the snapshot once its version is newer than the one given,
// at once if it already is. It returns false when ctx ends first.
func (s *Session) WaitSince(ctx context.Context, version int64) (snapshot.Snapshot, bool) {
	for {
		s.mu.Lock()
		if s.version > version {
			snap := snapshot.From(s.state, s.version)
			s.mu.Unlock()
			return snap, true
		}
		changed := s.changed
		s.mu.Unlock()

		select {
		case <-changed:
		case <-ctx.Done():
			return snapshot.Snapshot{}, false
		}
	}
}

// Close stops the engine and waits until it has finished and its last event is
// in the state. It is safe to call twice, and Add refuses afterwards.
func (s *Session) Close() {
	s.adding.Lock()
	defer s.adding.Unlock()
	s.closed = true

	s.mu.Lock()
	cancel, pumped, done := s.cancel, s.pumped, s.done
	s.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
	<-pumped
}

// bumpLocked moves the version and wakes everyone waiting for it. s.mu must be held.
func (s *Session) bumpLocked() {
	s.version++
	close(s.changed)
	s.changed = make(chan struct{})
}

// start builds the engine for the first URLs and sets it running. s.adding must be held.
func (s *Session) start(urls []string) error {
	ctx, cancel := context.WithCancel(s.root)
	eng, err := s.factory(ctx, urls)
	if err != nil {
		cancel()
		return err
	}
	pumped := make(chan struct{})
	done := make(chan struct{})

	s.mu.Lock()
	s.state = runstate.New(eng.Sources())
	s.eng, s.cancel, s.pumped, s.done = eng, cancel, pumped, done
	s.bumpLocked()
	s.mu.Unlock()

	go s.pump(eng, pumped)
	go func() {
		eng.Run(ctx)
		close(done)
	}()
	return nil
}

// pump folds the engine's events into the state, one version per event, until
// the engine closes its channel.
func (s *Session) pump(eng Engine, pumped chan struct{}) {
	for ev := range eng.Events() {
		s.mu.Lock()
		s.state.Apply(ev)
		s.bumpLocked()
		s.mu.Unlock()
	}
	close(pumped)
}

// engine returns the running engine, or nil when no URLs have arrived yet.
func (s *Session) engine() Engine {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eng
}
