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
	downloadpage.Chooser
	// Snapshot is every song the run has listed, with the version each is to download and its state.
	Snapshot() []engine.EntrySnapshot
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
	// PickFile shows the operating system's own open-file window, starting in start, and
	// returns the file chosen; chosen is false when the person cancelled. Without it the
	// path is asked for on the bottom line.
	PickFile func(start string) (path string, chosen bool, err error)
	// SaveFile is the same for saving: it starts in dir with name offered.
	SaveFile func(dir, name string) (path string, chosen bool, err error)
	// Open opens an address in the person's browser; nil for none.
	Open func(url string) error
	// OpenFolder shows a folder in the system's file manager; nil for none.
	OpenFolder func(path string) error
	// Quit is called when the person quits, so a host that shows the program in a window
	// can close it; nil for none.
	Quit func()
	// Clipboard is the machine's own clipboard, which the page around the terminal cannot
	// reach; nil for none, and the terminal's own is then asked for.
	Clipboard Clipboard
	// Tools says whether what a download needs (the downloader and the video tools) is there.
	// On the first start they are fetched while the window is already open, and until they are
	// the downloads cannot start; nil means they are always there.
	Tools func() ToolsState
	// Version is the version of the app, which the About window tells; empty for a build
	// that has none.
	Version string
}

// Clipboard is the text the person copied, and a way to copy.
type Clipboard interface {
	Read() (string, error)
	Write(text string) error
}

// ToolsState is whether the tools of the downloads are ready, and what to tell the person when
// they are not: that they are being fetched, or that one could not be.
type ToolsState struct {
	Ready   bool
	Message string
	// Failed is true when the message is about something that went wrong, not about waiting.
	Failed bool
}
