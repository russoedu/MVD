package appwindow

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// backdropEvent is the event the page listens to, to dim itself while a popup is open.
const backdropEvent = "mvd:modal"

// Backdrop tells the page when a popup is open over the terminal, so the page around the
// terminal can be dimmed with it. It is made before the window and works once the window is
// up; before that, and after, Set does nothing.
type Backdrop struct {
	mu  sync.Mutex
	app *application.App
}

func (b *Backdrop) attach(app *application.App) {
	b.mu.Lock()
	b.app = app
	b.mu.Unlock()
}

// Set says whether a popup is open.
func (b *Backdrop) Set(open bool) {
	b.mu.Lock()
	app := b.app
	b.mu.Unlock()
	if app != nil {
		app.Event.Emit(backdropEvent, open)
	}
}
