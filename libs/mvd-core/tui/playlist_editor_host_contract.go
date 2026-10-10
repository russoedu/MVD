package tui

import (
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
)

// PlanDownloadStarter starts a run that downloads exactly what a plan names.
type PlanDownloadStarter func(cfg config.Config, plan []engine.PlannedEntry) (Run, error)

// Planner is what a plan-only run offers once it has dealt with every song.
type Planner interface {
	Plan() []engine.PlannedEntry
}

// EditorHost is what the playlist editor needs from the machine it runs on. Every field
// is optional: without Plan the editor can only open files, without PlanDownload it can
// not download, and so on.
type EditorHost struct {
	// Plan starts a run that lists, names and picks, and downloads nothing. The run must
	// also be a Planner.
	Plan RunStarter
	// PlanDownload starts the download of a reviewed plan.
	PlanDownload PlanDownloadStarter
	// Open opens an address in the person's browser.
	Open func(url string) error
	// Describe returns the title of a YouTube video, so a video the person picked can be
	// named; it may be slow.
	Describe func(videoID string) (string, error)
	// SessionFile is where the review in progress is kept after every change, so it
	// can be brought back the next time.
	SessionFile string
	// SaveDir is the folder offered when the person saves the review as a .mvd file.
	SaveDir string
	// DecisionLog is the file the reasons of changed decisions are appended to; it is only
	// used in a build made for testing (see playlisteditor.DebugBuild).
	DecisionLog string
}
