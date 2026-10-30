// Package signin lets a person sign in to YouTube inside MVD: a window of its own, in a
// separate program run, keeps the login in a profile folder that yt-dlp can then read.
package signin

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Flag is the command line flag that makes the program show the sign-in window and nothing else.
const Flag = "youtube-sign-in"

// loginURL is Google's sign-in, which carries on to YouTube Music.
const loginURL = "https://accounts.google.com/ServiceLogin?service=youtube&continue=https%3A%2F%2Fmusic.youtube.com%2F"

// RunWindow shows the sign-in window with its browser profile in profileDir and returns when it
// is closed. It is its own program run, so the window has a profile apart from the main one, and
// nothing holds the cookie database once it has ended.
func RunWindow(profileDir string) error {
	app := application.New(application.Options{
		Name:    "MVD sign in",
		Windows: application.WindowsOptions{WebviewUserDataPath: profileDir},
		Mac:     application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Sign in to YouTube", Width: 520, Height: 760, URL: loginURL,
	})
	window.RegisterHook(events.Common.WindowClosing, func(*application.WindowEvent) { app.Quit() })

	return app.Run()
}
