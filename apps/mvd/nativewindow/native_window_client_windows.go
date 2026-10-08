//go:build windows

package nativewindow

import (
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	webview2 "github.com/jchv/go-webview2"
)

const (
	windowWidth  = 1100
	windowHeight = 760
	swRestore    = 9
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procShowWindow    = user32.NewProc("ShowWindow")
	procSetForeground = user32.NewProc("SetForegroundWindow")
	procIsIconic      = user32.NewProc("IsIconic")
	mu                sync.Mutex
	current           *window
)

// window is the one window that may be open.
type window struct {
	handle uintptr
	// closed is closed when the window is gone.
	closed chan struct{}
}

// Open shows url in a native window. Calling it while the window is open brings
// that window to the front instead of opening a second one. It returns once the
// window is up, or ErrUnavailable when WebView2 cannot start (not installed).
// The channel it returns is closed when the window is closed: a process that has
// nothing else to do waits on it, because the window goes with the process.
// dataDir is where the web view keeps its profile; it must be a folder the
// person can write to, never the program's own.
//
// The url is always one this program built (http on loopback), never user input.
func Open(url, dataDir string) (<-chan struct{}, error) {
	mu.Lock()
	if current != nil {
		raise(current.handle)
		closed := current.closed
		mu.Unlock()
		return closed, nil
	}
	win := &window{closed: make(chan struct{})}
	current = win
	mu.Unlock()

	started := make(chan error, 1)
	go run(win, url, filepath.Join(dataDir, "webview"), started)
	if err := <-started; err != nil {
		return nil, err
	}
	return win.closed, nil
}

// run owns the window: Windows delivers a window's messages to the thread that
// made it, so that thread stays this goroutine's until the window closes.
func run(win *window, url, profileDir string, started chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(win.closed)

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:  profileDir,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "MVD",
			Width:  windowWidth,
			Height: windowHeight,
			Center: true,
		},
	})
	if w == nil {
		forget()
		started <- ErrUnavailable
		return
	}
	defer w.Destroy()
	defer forget()

	mu.Lock()
	win.handle = uintptr(w.Window())
	mu.Unlock()

	w.Navigate(url)
	started <- nil
	w.Run()
}

func forget() {
	mu.Lock()
	current = nil
	mu.Unlock()
}

// raise brings an open window to the front, restoring it when minimised.
func raise(handle uintptr) {
	if handle == 0 {
		return
	}
	if minimised, _, _ := procIsIconic.Call(handle); minimised != 0 {
		_, _, _ = procShowWindow.Call(handle, swRestore)
	}
	_, _, _ = procSetForeground.Call(handle)
}
