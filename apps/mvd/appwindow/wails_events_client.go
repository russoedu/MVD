package appwindow

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// wailsEvents is Wails' event bus between the page and Go, in the shape tty-go binds a
// program to (ttygo.Events).
type wailsEvents struct {
	app *application.App
}

// On calls handler with the text each event called name carries from the page.
func (e wailsEvents) On(name string, handler func(data string)) func() {
	return e.app.Event.On(name, func(event *application.CustomEvent) {
		if text, ok := event.Data.(string); ok {
			handler(text)
		}
	})
}

// Emit sends the page an event called name.
func (e wailsEvents) Emit(name, data string) {
	e.app.Event.Emit(name, data)
}
