package downloadlist

import (
	"context"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// DefaultSaveEvery is how often Watch looks at the run: often enough that little is lost
// if the program dies, seldom enough that a long playlist is not rewritten all the time.
const DefaultSaveEvery = 2 * time.Second

// Watch keeps the list file up to date with the engine while it runs: every `every` it
// takes the engine's snapshot and writes the list if anything changed, and it writes once
// more when ctx ends. It returns after that last write. The list must have a file, or
// there is nothing to write to.
func (s *Session) Watch(ctx context.Context, eng *engine.Engine, every time.Duration) {
	if every <= 0 {
		every = DefaultSaveEvery
	}
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			_ = s.Sync(eng)
		case <-ctx.Done():
			_ = s.Sync(eng)
			return
		}
	}
}

// Sync writes the list now, from the engine as it stands.
func (s *Session) Sync(eng *engine.Engine) error {
	return s.SyncSnapshot(eng.Snapshot())
}

// SyncSnapshot writes the list now, from a snapshot of a run.
func (s *Session) SyncSnapshot(snapshot []engine.EntrySnapshot) error {
	s.Apply(snapshot)
	return s.flush()
}

// Decide puts the decisions of a review on the songs they belong to, by key: a song with
// no decision there keeps what it has. It writes the list when it has a file.
func (s *Session) Decide(decided []playlistfile.Entry) error {
	s.mu.Lock()
	byKey := make(map[string]playlistfile.Entry, len(decided))
	for _, d := range decided {
		byKey[d.Key()] = d
	}
	for i := range s.file.Items {
		songs := s.file.Items[i].Songs
		for j := range songs {
			if d, ok := byKey[songs[j].Key()]; ok {
				songs[j].Decision, songs[j].ChosenID, songs[j].ChosenTitle = d.Decision, d.ChosenID, d.ChosenTitle
			}
		}
	}
	s.mu.Unlock()
	return s.flush()
}

// Replace puts the given items in the list, as the person edited them, and writes the
// list when it has a file.
func (s *Session) Replace(items []playlistfile.Item) error {
	s.mu.Lock()
	s.file.Items = items
	s.mu.Unlock()
	return s.flush()
}
