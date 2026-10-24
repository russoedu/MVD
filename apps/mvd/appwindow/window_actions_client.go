package appwindow

import (
	"context"
	"sync/atomic"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// windowEvent is the event the buttons of the page's title bar send.
const windowEvent = "mvd:window"

// doWindowAction does what a button of the title bar asked: the window has no frame of the
// system's, so the page asks for minimise, maximise and close. Close asks first when
// downloads are running, as closing a window with a frame would.
func doWindowAction(ctx context.Context, action string, busy func() bool, quitting *atomic.Bool) {
	switch action {
	case "minimise":
		wailsruntime.WindowMinimise(ctx)
	case "maximise":
		wailsruntime.WindowToggleMaximise(ctx)
	case "close":
		if !keepOpen(busy, func() bool { return askToQuit(ctx) }) {
			// Quitting closes the window, which would ask again: it has been answered.
			quitting.Store(true)
			wailsruntime.Quit(ctx)
		}
	}
}
