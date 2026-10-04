package tray

import "context"

// serveTrayMenu reacts to the tray menu until the tray should go away, then calls
// exit exactly once and returns.
//
// "Open MVD" calls open, as often as it is clicked. "Quit" calls quit (which stops the
// app) and then exit (which removes the icon). If ctx ends first (Ctrl+C, or the server
// stopped), only exit is called. exit must always run: the tray blocks the main
// goroutine until it does, so skipping it on Quit leaves the app running with its
// server still listening, whatever quit did.
func serveTrayMenu(ctx context.Context, openClicked, quitClicked <-chan struct{}, open, quit, exit func()) {
	defer exit()
	for {
		select {
		case <-openClicked:
			open()
		case <-quitClicked:
			quit()

			return
		case <-ctx.Done():
			return
		}
	}
}
