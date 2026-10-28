package appwindow

import (
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// preferencesURL is the page of the preferences window: the same page as MVD's window, which
// shows only the preferences when it is told so (see the page's main.tsx).
const preferencesURL = "/?window=preferences"

// The size of the preferences window, and the smallest it can be made.
const (
	preferencesWidth, preferencesHeight       = 820, 680
	preferencesMinWidth, preferencesMinHeight = 520, 380
)

// preferencesWindow is the window of its own the preferences open in, over MVD's window. The
// program says when they are open (see ClassicWindow.ShowPreferences); the person can also close
// the window with the system's own means, which is told to the program as the close button of the
// page would be.
type preferencesWindow struct {
	app *application.App
	// main is MVD's window, which the preferences open over and which they send the focus to.
	main *application.WebviewWindow
	// onClose is told when the person closes the window; the program then decides.
	onClose func()

	mu        sync.Mutex
	window    *application.WebviewWindow
	byProgram *atomic.Bool // the program is closing this window, so it closes without asking
}

// set opens the window when the program has the preferences open and closes it when it has not.
func (p *preferencesWindow) set(open bool) {
	// The window's own calls wait for the system's thread, which may be waiting for focus() here
	// (MVD's window gets the focus back when this one closes), so the lock is never held across them.
	p.mu.Lock()
	window, byProgram := p.window, p.byProgram
	if !open {
		p.window, p.byProgram = nil, nil
	}
	p.mu.Unlock()

	switch {
	case open && window == nil:
		window, byProgram = p.create()
		p.mu.Lock()
		p.window, p.byProgram = window, byProgram
		p.mu.Unlock()
	case open:
		window.Focus()
	case window != nil:
		byProgram.Store(true)
		window.Close()
	}
}

// focus brings the window to the front, when it is open: MVD's window cannot be used while the
// preferences are open, so a click on it leads back here.
func (p *preferencesWindow) focus() {
	p.mu.Lock()
	window := p.window
	p.mu.Unlock()
	if window != nil {
		window.Focus()
	}
}

// create opens the window, in the middle of MVD's window.
func (p *preferencesWindow) create() (*application.WebviewWindow, *atomic.Bool) {
	x, y := 0, 0
	if p.main != nil {
		left, top := p.main.Position()
		width, height := p.main.Size()
		x = max(left+(width-preferencesWidth)/2, 0)
		y = max(top+(height-preferencesHeight)/2, 0)
	}
	window := p.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      "preferences",
		Title:     "MVD Preferences",
		Width:     preferencesWidth,
		Height:    preferencesHeight,
		MinWidth:  preferencesMinWidth,
		MinHeight: preferencesMinHeight,
		// The page draws the title bar, in the style of the window, as MVD's window does.
		Frameless:        true,
		BackgroundColour: application.NewRGB(0, 0, 0),
		URL:              preferencesURL,
		InitialPosition:  application.WindowXY,
		X:                x,
		Y:                y,
	})
	byProgram := &atomic.Bool{}
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if byProgram.Load() {
			return
		}
		event.Cancel() // the program decides, and closes the window when it has closed the preferences
		if p.onClose != nil {
			p.onClose()
		}
	})
	window.Focus()

	return window, byProgram
}
