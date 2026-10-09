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

	ttygo "github.com/meta-tui/treactui/packages/tty-go"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
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
	// Busy reports whether downloads are running, so closing the window asks first.
	Busy func() bool
}

// Run opens the window and returns when it is closed. A second start of the app does not
// open a second window: it raises the first one and ends. Cancelling ctx closes the window.
func Run(ctx context.Context, opts Options) error {
	var windowCtx context.Context
	var unbind func()

	err := wails.Run(&options.App{
		Title:            "MVD",
		Width:            1100,
		Height:           760,
		MinWidth:         640,
		MinHeight:        400,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 255},
		AssetServer:      &assetserver.Options{Assets: pageAssets()},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: uniqueID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				if windowCtx == nil {
					return
				}
				wailsruntime.WindowUnminimise(windowCtx)
				wailsruntime.Show(windowCtx)
			},
		},
		OnStartup: func(c context.Context) {
			windowCtx = c
			unbind = ttygo.BindShared(c, opts.Program, wailsEvents{ctx: c}, ttygo.BindOptions{})
			go func() {
				<-ctx.Done()
				wailsruntime.Quit(c)
			}()
		},
		OnShutdown: func(context.Context) {
			if unbind != nil {
				unbind()
			}
		},
		OnBeforeClose: func(c context.Context) bool {
			return keepOpen(opts.Busy, func() bool { return askToQuit(c) })
		},
		Menu: appMenu(),
		Windows: &windows.Options{
			WebviewUserDataPath: filepath.Join(opts.DataDir, "webview"),
		},
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

// askToQuit asks whether to close the app although downloads are running.
func askToQuit(ctx context.Context) bool {
	answer, err := wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
		Type:          wailsruntime.QuestionDialog,
		Title:         "MVD",
		Message:       "Downloads are still running. Quit and stop them?",
		Buttons:       []string{"Quit", "Keep downloading"},
		DefaultButton: "Keep downloading",
		CancelButton:  "Keep downloading",
	})
	// Windows has its own Yes and No buttons and answers with those; the other systems
	// answer with the label of the button.
	return err == nil && (answer == "Quit" || answer == "Yes")
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
