package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
)

// fakeEngine stands in for the real engine: it runs until cancelled and sends
// only the events a test pushes into it.
type fakeEngine struct {
	mu        sync.Mutex
	sources   []engine.PlaylistSource
	events    chan interface{}
	refuse    bool
	retried   []int
	playlists []int
}

func newFakeEngine(urls ...string) *fakeEngine {
	f := &fakeEngine{events: make(chan interface{}, 256)}
	for i, u := range urls {
		f.sources = append(f.sources, engine.PlaylistSource{Index: i, URL: u})
	}
	return f
}

func (f *fakeEngine) Run(ctx context.Context) {
	<-ctx.Done()
	close(f.events)
}

func (f *fakeEngine) Events() <-chan interface{} { return f.events }

func (f *fakeEngine) Sources() []engine.PlaylistSource {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]engine.PlaylistSource(nil), f.sources...)
}

func (f *fakeEngine) AddSource(url string) (engine.PlaylistSource, bool) {
	f.mu.Lock()
	if f.refuse {
		f.mu.Unlock()
		return engine.PlaylistSource{}, false
	}
	src := engine.PlaylistSource{Index: len(f.sources), URL: url}
	f.sources = append(f.sources, src)
	f.mu.Unlock()
	f.events <- engine.EvPlaylistAdded{Source: src}
	return src, true
}

func (f *fakeEngine) Retry(entryID int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.retried = append(f.retried, entryID)
	return entryID == 1
}

func (f *fakeEngine) RetryPlaylist(playlist int) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playlists = append(f.playlists, playlist)
	return 2
}

func (f *fakeEngine) emit(ev interface{}) { f.events <- ev }

func (f *fakeEngine) sourceURLs() []string {
	var out []string
	for _, s := range f.Sources() {
		out = append(out, s.URL)
	}
	return out
}

// engines hands out one fake engine per factory call and remembers the calls.
type engines struct {
	mu    sync.Mutex
	calls [][]string
	built []*fakeEngine
	fail  error
}

func (e *engines) factory(_ context.Context, urls []string) (Engine, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls = append(e.calls, append([]string(nil), urls...))
	if e.fail != nil {
		err := e.fail
		e.fail = nil
		return nil, err
	}
	f := newFakeEngine(urls...)
	e.built = append(e.built, f)
	return f, nil
}

func (e *engines) engine(t *testing.T) *fakeEngine {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.built) == 0 {
		t.Fatal("no engine was built")
	}
	return e.built[len(e.built)-1]
}

// newSession returns a session over fake engines, closed when the test ends.
func newSession(t *testing.T) (*Session, *engines) {
	t.Helper()
	fakes := &engines{}
	s := New(context.Background(), fakes.factory)
	t.Cleanup(s.Close)
	return s, fakes
}

// waitFor polls until the condition holds: the events pump applies events a
// moment after they are sent.
func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

var errBuild = errors.New("cannot build the engine")
