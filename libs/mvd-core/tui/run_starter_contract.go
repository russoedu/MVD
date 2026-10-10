package tui

import "youtube-downloader/libs/mvd-core/config"

// Run is one download run a host started for the app model.
type Run interface {
	Controller
	// Close stops the run, drains its events and waits for it to wind down.
	Close()
}

// RunStarter starts a download run for the chosen settings and list. The app
// model calls it when the user starts from the setup screens; an error is
// shown on the setup screen and the user stays there.
type RunStarter func(cfg config.Config, urls []string) (Run, error)

// AppInput seeds an app model.
type AppInput struct {
	Setup SetupInput
	Start RunStarter
	// Editor is what the playlist editor needs; its zero value offers an editor that can
	// only open files.
	Editor EditorHost
	// ReviewOnStart opens the playlist editor on the list at once, instead of the list.
	ReviewOnStart bool
}
