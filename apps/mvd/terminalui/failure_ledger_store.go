package terminalui

import (
	"sync"

	"youtube-downloader/libs/mvd-core/engine"
)

// retrier is the part of the engine a retry needs.
type retrier interface {
	Retry(entryID int) bool
	AddSource(url string) (engine.PlaylistSource, bool)
}

// failureLedger remembers what has failed in a run, from its events: the entries
// that failed and the links that could not be looked up. The download screen keeps
// the same facts for itself; the host needs them to try again once something that
// was behind the failures (an out-of-date downloader) has been fixed.
type failureLedger struct {
	mu        sync.Mutex
	entries   map[int]struct{}
	playlists map[int]string // playlist index to its link
	linkOf    func(playlist int) string
}

func newFailureLedger(linkOf func(playlist int) string) *failureLedger {
	return &failureLedger{entries: map[int]struct{}{}, playlists: map[int]string{}, linkOf: linkOf}
}

// apply folds one event in and says whether it reports a failure: a download that
// failed, or a link that could not even be looked up, which is how a downloader
// that has gone out of date usually shows itself.
func (l *failureLedger) apply(ev interface{}) (failure bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch e := ev.(type) {
	case engine.EvEntryState:
		if e.State == engine.StateFailed {
			l.entries[e.Entry] = struct{}{}
			return true
		}
		delete(l.entries, e.Entry)
	case engine.EvPlaylistFailed:
		l.playlists[e.Playlist] = l.linkOf(e.Playlist)
		return true
	case engine.EvPlaylistListed:
		delete(l.playlists, e.Playlist)
	}
	return false
}

// retryAll queues everything that has failed again and says how many it queued: every
// failed entry, and every link that could not be looked up, added once more.
func (l *failureLedger) retryAll(r retrier) int {
	l.mu.Lock()
	entries := make([]int, 0, len(l.entries))
	for id := range l.entries {
		entries = append(entries, id)
	}
	links := make(map[int]string, len(l.playlists))
	for index, link := range l.playlists {
		links[index] = link
	}
	l.mu.Unlock()

	retried := 0
	for _, id := range entries {
		if r.Retry(id) {
			retried++
		}
	}
	for index, link := range links {
		if _, queued := r.AddSource(link); queued {
			retried++
			// The new copy is the one to watch from here on; this one is dealt with.
			l.mu.Lock()
			delete(l.playlists, index)
			l.mu.Unlock()
		}
	}
	return retried
}
