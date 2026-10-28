// Package appwindow shows the app in a window of its own, drawn by the system's web view
// (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux) through Wails 3. Wails serves
// the page itself and carries the terminal interface between the page and Go as events, so
// there is no server and no port. The app is a normal program: closing the main window
// quits.
package appwindow

import (
	"context"
	_ "embed"
	"errors"
	"path/filepath"
	"runtime"
	"sync/atomic"

	ttygo "github.com/meta-ui-labs/treactui/packages/tty-go"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed window_icon.png
var windowIcon []byte

// uniqueID names this program to the system, so a second start finds the first.
const uniqueID = "io.github.russoedu.mvd"

// Options says what the window shows.
type Options struct {
	// Program is the terminal interface the window shows.
	Program *ttygo.SharedProgram
	// DataDir is the app-data folder; the web view keeps its profile in a folder inside.
	DataDir string
	// OnCloseRequest is told when the person asks to close the window (its close button, or the
	// system's): the window stays open and the program decides, asking first when songs download.
	// Nil closes the window at once.
	OnCloseRequest func()
	// OnSecondStart is told the arguments of a second start of the app, which raises this
	// window instead of opening its own: a .mvd file double clicked while the app is open
	// arrives this way. Nil for none.
	OnSecondStart func(args []string)
	// OnFileOpen is told the path of a file the system opens with this app. macOS does not put
	// a double clicked file in the arguments, in a first start or a second one: it sends this
	// instead. Nil for none.
	OnFileOpen func(path string)
	// Backdrop, when set, is told by the program when a popup is open, and tells the page.
	Backdrop *Backdrop
	// Classic, when set, carries the classic themed window between the program and the page.
	Classic *ClassicWindow
}

// Run opens the window and returns when it is closed. A second start of the app does not
// open a second window: it raises the first one and ends. Cancelling ctx closes the window.
func Run(ctx context.Context, opts Options) error {
	var window *application.WebviewWindow
	var quitting atomic.Bool // the close button has asked, and been answered

	app := application.New(application.Options{
		Name:   "MVD",
		Icon:   windowIcon,
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(pageAssets())},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: uniqueID,
			OnSecondInstanceLaunch: func(second application.SecondInstanceData) {
				if window == nil {
					return
				}
				if opts.OnSecondStart != nil {
					opts.OnSecondStart(second.Args)
				}
				window.UnMinimise()
				window.Show()
				window.Focus()
			},
		},
		Mac:     application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
		Windows: application.WindowsOptions{WebviewUserDataPath: filepath.Join(opts.DataDir, "webview")},
		Linux:   application.LinuxOptions{ProgramName: "MVD"},
	})

	bus := wailsEvents{app: app}
	stopWindowEvents := app.Event.On(windowEvent, func(event *application.CustomEvent) {
		if action, ok := event.Data.(string); ok && window != nil {
			doWindowAction(app, window, action, opts.OnCloseRequest, &quitting)
		}
	})
	var stopClassic func()
	if opts.Backdrop != nil {
		opts.Backdrop.attach(app)
	}
	if opts.Classic != nil {
		stopClassic = opts.Classic.attach(app)
	}
	if opts.OnFileOpen != nil {
		app.Event.OnApplicationEvent(events.Common.ApplicationOpenedWithFile, func(event *application.ApplicationEvent) {
			if path := event.Context().Filename(); path != "" {
				opts.OnFileOpen(path)
			}
		})
	}
	unbind := ttygo.BindShared(ctx, opts.Program, bus, ttygo.BindOptions{})

	window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      "main",
		Title:     "MVD",
		Width:     1100,
		Height:    760,
		MinWidth:  640,
		MinHeight: 400,
		// The page draws the title bar and the buttons of the window (see the page header).
		Frameless:        true,
		BackgroundColour: application.NewRGB(0, 0, 0),
		URL:              "/",
	})
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if quitting.Load() || opts.OnCloseRequest == nil {
			return
		}
		event.Cancel() // the program decides, asking first when songs download
		opts.OnCloseRequest()
	})

	go func() {
		<-ctx.Done()
		quitting.Store(true) // the program has decided: the window closes without asking again
		app.Quit()
	}()

	if runtime.GOOS == "darwin" {
		app.Menu.Set(appMenu())
	}
	err := app.Run()
	stopWindowEvents()
	if stopClassic != nil {
		stopClassic()
	}
	unbind()
	if err != nil {
		return errors.New("cannot open the app window: " + err.Error())
	}

	return nil
}

// appMenu is the menu bar of macOS, which has the Quit, Copy and Paste of every app; other
// systems have no menu, as the page draws everything.
func appMenu() *application.Menu {
	menu := application.NewMenu()
	menu.AddRole(application.AppMenu)
	menu.AddRole(application.EditMenu)
	menu.AddRole(application.WindowMenu)

	return menu
}
