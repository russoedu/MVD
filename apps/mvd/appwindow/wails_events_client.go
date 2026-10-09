package appwindow

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// wailsEvents is Wails' event bus between the page and Go, in the shape tty-go binds a
// program to (ttygo.Events).
type wailsEvents struct {
	ctx context.Context
}

// On calls handler with the text each event called name carries from the page.
func (e wailsEvents) On(name string, handler func(data string)) func() {
	return wailsruntime.EventsOn(e.ctx, name, func(data ...interface{}) {
		if len(data) == 0 {
			return
		}
		if text, ok := data[0].(string); ok {
			handler(text)
		}
	})
}

// Emit sends the page an event called name.
func (e wailsEvents) Emit(name, data string) {
	wailsruntime.EventsEmit(e.ctx, name, data)
}
