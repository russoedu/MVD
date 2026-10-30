// Package signin lets a person sign in to YouTube inside MVD: a window of its own, in a
// separate program run, keeps the login in a profile folder that yt-dlp can then read.
package signin

import (
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Flag is the command line flag that makes the program show the sign-in window and nothing else.
const Flag = "youtube-sign-in"

// loginURL is Google's sign-in, which carries on to YouTube Music.
const loginURL = "https://accounts.google.com/ServiceLogin?service=youtube&continue=https%3A%2F%2Fmusic.youtube.com%2F"

// signedInEvent is what the page sends when YouTube Music has opened.
const signedInEvent = "mvd-signed-in"

// doneScript runs after every page of the window: on YouTube Music it tells the program.
const doneScript = `if (location.hostname === "music.youtube.com") { window.chrome.webview.postMessage("wails:event:emit:` + signedInEvent + `"); }`

// closeDelay lets the page finish setting its cookies before the window closes.
const closeDelay = 4 * time.Second

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
		Title: "Sign in to YouTube - this window closes by itself once YouTube Music has opened",
		Width: 520, Height: 760, URL: loginURL,
		// The page tells the program when the sign-in is done (see doneScript).
		AllowSimpleEventEmit: true,
	})
	window.RegisterHook(events.Common.WindowClosing, func(*application.WindowEvent) { app.Quit() })
	// Google carries on to YouTube Music once the person has signed in. A moment later the window
	// closes, which also writes the login to the profile folder.
	app.Event.On(signedInEvent, func(*application.CustomEvent) {
		go func() {
			time.Sleep(closeDelay)
			app.Quit()
		}()
	})
	window.OnWindowEvent(events.Windows.WebViewNavigationCompleted, func(*application.WindowEvent) {
		// Wails holds scripts back until the page says its own runtime is ready, which a page of
		// Google never does.
		window.HandleMessage("wails:runtime:ready")
		window.ExecJS(doneScript)
	})

	return app.Run()
}
