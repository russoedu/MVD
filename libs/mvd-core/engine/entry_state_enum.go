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
	// StateNotFound is a song with no official video, when only official videos are
	// wanted: it is not downloaded.
	StateNotFound
	// StatePlanned is a song a plan-only run has chosen the version of: it is not
	// downloaded, the plan says what would be.
	StatePlanned
	// StateSkipped is a song the person cancelled while it waited for their choice: it is not
	// downloaded.
	StateSkipped
	// StateAccepted is a song the person accepted: it waits in the queue for the downloads to start,
	// and shows as downloading once a downloader takes it.
	StateAccepted
)

// String returns the lower-case name of the state, or "unknown" for a value that has none.
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
	case StateNotFound:
		return "not found"
	case StatePlanned:
		return "planned"
	case StateSkipped:
		return "skipped"
	case StateAccepted:
		return "accepted"
	}
	return "unknown"
}

// Finished reports whether the state is terminal.
func (s EntryState) Finished() bool {
	return s == StateDone || s == StateDuplicate || s == StateFailed || s == StateNotFound || s == StatePlanned || s == StateSkipped
}
