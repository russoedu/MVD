// Package appwindow shows the app in a window of its own, drawn by the system's web view
// (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux) through Wails. Wails serves
// the page itself and carries the terminal interface between the page and Go as events, so
// there is no server and no port. The app is a normal program: one window, and closing it
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
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	var windowCtx context.Context
	var unbind, stopWindowEvents, stopClassic func()
	var quitting atomic.Bool // the close button has asked, and been answered

	err := wails.Run(&options.App{
		Title:            "MVD",
		Width:            1100,
		Height:           760,
		MinWidth:         640,
		MinHeight:        400,
		// The page draws the title bar and the buttons of the window (see the page header).
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 255},
		AssetServer:      &assetserver.Options{Assets: pageAssets()},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: uniqueID,
			OnSecondInstanceLaunch: func(second options.SecondInstanceData) {
				if windowCtx == nil {
					return
				}
				if opts.OnSecondStart != nil {
					opts.OnSecondStart(second.Args)
				}
				wailsruntime.WindowUnminimise(windowCtx)
				wailsruntime.Show(windowCtx)
			},
		},
		OnStartup: func(c context.Context) {
			windowCtx = c
			stopWindowEvents = wailsruntime.EventsOn(c, windowEvent, func(data ...interface{}) {
				if len(data) > 0 {
					if action, ok := data[0].(string); ok {
						doWindowAction(c, action, opts.OnCloseRequest, &quitting)
					}
				}
			})
			if opts.Backdrop != nil {
				opts.Backdrop.attach(c)
			}
			if opts.Classic != nil {
				stopClassic = opts.Classic.attach(c)
			}
			unbind = ttygo.BindShared(c, opts.Program, wailsEvents{ctx: c}, ttygo.BindOptions{})
			go func() {
				<-ctx.Done()
				quitting.Store(true) // the program has decided: the window closes without asking again
				wailsruntime.Quit(c)
			}()
		},
		OnShutdown: func(context.Context) {
			if stopWindowEvents != nil {
				stopWindowEvents()
			}
			if stopClassic != nil {
				stopClassic()
			}
			if unbind != nil {
				unbind()
			}
		},
		OnBeforeClose: func(c context.Context) bool {
			if quitting.Load() {
				return false
			}
			if opts.OnCloseRequest == nil {
				return false
			}
			opts.OnCloseRequest()

			return true
		},
		Menu: appMenu(),
		Windows: &windows.Options{
			WebviewUserDataPath: filepath.Join(opts.DataDir, "webview"),
		},
		Mac: &mac.Options{OnFileOpen: opts.OnFileOpen},
		Linux: &linux.Options{
			Icon:        windowIcon,
			ProgramName: "MVD",
		},
	})
	if err != nil {
		return errors.New("cannot open the app window: " + err.Error())
	}
	return nil
}

// appMenu is the menu bar macOS expects of an app: the app menu (with Quit) and the edit
// menu, whose items are what make Cmd+C, Cmd+V and the like work in the page. Windows and
// Linux have no menu bar.
func appMenu() *menu.Menu {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return menu.NewMenuFromItems(menu.AppMenu(), menu.EditMenu(), menu.WindowMenu())
}
