// Package engine downloads every playlist: it lists them, resolves official
// videos, runs yt-dlp per entry from a bounded worker pool and reports all
// observable state as events.
package engine

// EntryState is the life cycle of one playlist entry.
type EntryState int

const (
	StateQueued EntryState = iota
	StateResolving
	StateDownloading
	StateMerging
	StateDone
	StateDuplicate
	StateFailed
)

func (s EntryState) String() string {
	switch s {
	case StateQueued:
		return "queued"
	case StateResolving:
		return "resolving"
	case StateDownloading:
		return "downloading"
	case StateMerging:
		return "merging"
	case StateDone:
		return "done"
	case StateDuplicate:
		return "duplicate"
	case StateFailed:
		return "failed"
	}
	return "unknown"
}

// Finished reports whether the state is terminal.
func (s EntryState) Finished() bool {
	return s == StateDone || s == StateDuplicate || s == StateFailed
}
