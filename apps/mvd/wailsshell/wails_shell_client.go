// Package wailsshell shows the app's page in a Wails v3 window with a system
// tray icon: the experiment that would replace the go-webview2 window and the
// fyne tray with one framework.
package wailsshell

import (
	"context"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func init() {
	// The window and the tray icon are drawn from the main thread (macOS needs it).
	runtime.LockOSThread()
}

// Run opens the window on url (when showAtStart), puts an icon in the system
// tray and blocks until the app quits. "Open MVD" and a click on the icon show
// the window; closing the window only hides it, so downloads keep running.
// "Quit" and the end of ctx quit the app, and Quit calls quit first.
//
// It must be called on the main goroutine. dataDir is where the web view keeps
// its profile.
func Run(ctx context.Context, url, dataDir string, icon []byte, showAtStart bool, quit func()) error {
	app := application.New(application.Options{
		Name:        "MVD",
		Description: "Music Video Downloader",
		// The page is served by the app's own local server; nothing is bundled here.
		Assets: application.AssetOptions{Handler: http.NotFoundHandler()},
		Windows: application.WindowsOptions{
			WebviewUserDataPath:           filepath.Join(dataDir, "webview"),
			DisableQuitOnLastWindowClosed: true,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:   "main",
		Title:  "MVD",
		Width:  1100,
		Height: 760,
		URL:    url,
		Hidden: !showAtStart,
	})
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	show := func() {
		window.Show()
		window.Focus()
	}

	menu := app.NewMenu()
	menu.Add("Open MVD").OnClick(func(*application.Context) { show() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) {
		quit()
		app.Quit()
	})

	tray := app.SystemTray.New()
	tray.SetIcon(icon)
	tray.SetLabel("MVD")
	tray.SetMenu(menu)
	tray.OnClick(show)

	go func() {
		<-ctx.Done()
		app.Quit()
	}()

	return app.Run()
}
