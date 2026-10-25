package review

import "youtube-downloader/libs/mvd-core/engine"

// Live is a run that holds each song for review (engine.Options.HoldForReview): the
// editor shows its songs as they are looked up, and a song the person has dealt with is
// sent to download at once, while the others are still being looked up or looked at.
type Live interface {
	// Snapshot is every song the run has listed, in the order it listed them.
	Snapshot() []engine.EntrySnapshot
	// Release sends song entry to download: as proposed when videoID is "", else that video.
	Release(entry int, videoID string) bool
	// ReleaseAll sends every song that is held, as proposed.
	ReleaseAll() int
}

// Planner is what a plan-only run offers once it has dealt with every song.
type Planner interface {
	Plan() []engine.PlannedEntry
}

// PlanRun is the plan-only run the editor waits for.
type PlanRun interface {
	Sources() []engine.PlaylistSource
}

// Options is what the editor needs to know of the machine it runs on. The can fields say what the
// host is able to do; without CanOpen the address is only shown, without CanDownload the songs
// cannot be downloaded, and so on.
type Options struct {
	// SessionFile is where the review in progress is kept after every change.
	SessionFile string
	// SaveDir is the folder offered when the person saves the review as a .mvd file.
	SaveDir string
	// DecisionLog is the file the reasons of changed decisions are appended to; it is only used in
	// a build made for testing (see playlisteditor.DebugBuild).
	DecisionLog string

	CanOpen     bool
	CanDescribe bool
	CanPickFile bool
	CanSaveFile bool
	CanDownload bool
}
