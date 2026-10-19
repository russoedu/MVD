package workspace

import (
	"youtube-downloader/libs/mvd-core/config"
	editorpage "youtube-downloader/libs/mvd-core/editor/editor"
	"youtube-downloader/libs/mvd-core/engine"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
)

// ListRun is a run that carries a download list: it can be paused, its failed songs can be
// tried again, and when it holds songs for review each can be released.
type ListRun interface {
	downloadpage.Run
	editorpage.Live
	Pause()
	Resume()
	Paused() bool
	RetryAll() int
	Reconfigure(engine.LiveSettings)
}

// Host is what the workspace needs from the machine it runs on.
type Host struct {
	// Start starts a run for the list: sources are the addresses nobody has looked at yet,
	// plan the songs of the others. With hold, every song waits for review until released.
	Start func(cfg config.Config, sources []string, plan []engine.PlannedEntry, hold bool) (ListRun, error)
	// Editor is what the review of a list needs.
	Editor editorpage.EditorHost
	// ListDir is the folder lists are opened from and saved to by default.
	ListDir string
	// Open opens an address in the person's browser; nil for none.
	Open func(url string) error
	// Quit is called when the person quits, so a host that shows the program in a window
	// can close it; nil for none.
	Quit func()
}
