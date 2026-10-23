package appwindow

import (
	"context"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// backdropEvent is the event the page listens to, to dim itself while a popup is open.
const backdropEvent = "mvd:modal"

// Backdrop tells the page when a popup is open over the terminal, so the page around the
// terminal can be dimmed with it. It is made before the window and works once the window is
// up; before that, and after, Set does nothing.
type Backdrop struct {
	mu  sync.Mutex
	ctx context.Context
}

func (b *Backdrop) attach(ctx context.Context) {
	b.mu.Lock()
	b.ctx = ctx
	b.mu.Unlock()
}

// Set says whether a popup is open.
func (b *Backdrop) Set(open bool) {
	b.mu.Lock()
	ctx := b.ctx
	b.mu.Unlock()
	if ctx != nil {
		wailsruntime.EventsEmit(ctx, backdropEvent, open)
	}
}
