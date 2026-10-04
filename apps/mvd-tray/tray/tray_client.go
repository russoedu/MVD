package tray

import (
	"context"
	"runtime"

	"fyne.io/systray"
)

func init() {
	// macOS draws the tray icon from the main thread, and Run blocks on it.
	runtime.LockOSThread()
}

const (
	openLabel = "Open MVD"
	quitLabel = "Quit"
)

// Run puts an icon in the system tray and blocks until it is gone. Clicking the
// icon or "Open MVD" opens url; "Quit" calls quit and removes the icon. When ctx ends
// (Ctrl+C, or the server stopped) the icon is removed and Run returns. If the tray
// cannot start (no desktop, no tray host) it returns at once without calling quit, and
// the caller carries on without an icon.
//
// It must be called on the main goroutine.
func Run(ctx context.Context, url string, open func(string) error, quit func()) {
	systray.Run(func() {
		systray.SetIcon(trayIcon(runtime.GOOS))
		systray.SetTitle("MVD")
		systray.SetTooltip("MVD - " + url)
		openPage := func() { _ = open(url) }
		systray.SetOnTapped(openPage)

		openItem := systray.AddMenuItem(openLabel, "Open MVD in your browser")
		systray.AddSeparator()
		quitItem := systray.AddMenuItem(quitLabel, "Stop MVD and its downloads")

		go serveTrayMenu(ctx, openItem.ClickedCh, quitItem.ClickedCh, openPage, quit, systray.Quit)
	}, func() {})
}
