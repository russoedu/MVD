package downloadlist

import (
	"context"
	"time"

	"youtube-downloader/libs/mvd-core/engine"
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
	s.Apply(eng.Snapshot())
	return s.flush()
}
