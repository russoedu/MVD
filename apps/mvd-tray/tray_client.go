package main

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

// runTray puts an icon in the system tray and blocks until it is gone. Clicking the
// icon or "Open MVD" opens url; "Quit" calls quit. When ctx ends (Ctrl+C, or the
// server stopped) the icon is removed and runTray returns. If the tray cannot start
// (no desktop, no tray host) it returns at once without calling quit, and the caller
// carries on without an icon.
//
// It must be called on the main goroutine.
func runTray(ctx context.Context, url string, open func(string) error, quit func()) {
	systray.Run(func() {
		systray.SetIcon(trayIcon(runtime.GOOS))
		systray.SetTitle("MVD")
		systray.SetTooltip("MVD - " + url)
		systray.SetOnTapped(func() { _ = open(url) })

		openItem := systray.AddMenuItem(openLabel, "Open MVD in your browser")
		systray.AddSeparator()
		quitItem := systray.AddMenuItem(quitLabel, "Stop MVD and its downloads")

		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					_ = open(url)
				case <-quitItem.ClickedCh:
					quit()

					return
				case <-ctx.Done():
					systray.Quit()

					return
				}
			}
		}()
	}, func() {})
}
