// MVD desktop app: a window of its own that shows the terminal interface (the same
// screens as the terminal app), served on localhost, and that takes new URLs while it
// downloads. Closing the window quits. The terminal app is apps/mvd-tui; both share
// libs/mvd-core.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"youtube-downloader/apps/mvd/appwindow"
	"youtube-downloader/apps/mvd/console"
	"youtube-downloader/apps/mvd/folderdialog"
	"youtube-downloader/apps/mvd/install"
	"youtube-downloader/apps/mvd/localserver"
	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/terminalui"
	"youtube-downloader/apps/mvd/toolupdates"
	"youtube-downloader/apps/mvd/uninstall"
	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/deps"
)

var version = "dev"

func main() {
	console.AttachParent()

	address := flag.String("addr", localserver.DefaultAddress, "address to serve the UI on; keep it on 127.0.0.1")
	noWindow := flag.Bool("no-window", false, "do not open the window: serve the interface until Ctrl+C")
	// Older shortcuts and scripts pass these two; they now mean the same as -no-window.
	noBrowser := flag.Bool("no-browser", false, "same as -no-window (kept for older scripts)")
	noTray := flag.Bool("no-tray", false, "same as -no-window (kept for older scripts)")
	movedFrom := flag.String("moved-from", "", "set by the app itself after moving to its folder: the old copy to remove")
	removeApp := flag.Bool("uninstall", false, "remove MVD from this computer, after asking; Settings > Apps on Windows runs this")
	flag.Parse()

	if *removeApp {
		if err := runUninstall(*address); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			console.ShowFatal(err.Error())
			os.Exit(1)
		}

		return
	}

	withWindow := !*noWindow && !*noBrowser && !*noTray
	if err := run(*address, withWindow, *movedFrom); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		console.ShowFatal(err.Error())
		os.Exit(1)
	}
}

func run(address string, withWindow bool, movedFrom string) error {
	logf := func(format string, a ...interface{}) { fmt.Printf(format+"\n", a...) }

	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}

	// The first time it is started from somewhere it does not belong, it offers to move
	// itself, and if that is accepted the moved copy takes over and this one is done.
	if install.OfferMoveHere(appDir, withWindow, movedFrom, version) {
		return nil
	}
	if movedFrom != "" {
		go install.CleanUpMovedProgram(movedFrom)
	}

	// The tools live in the app-data folder: the same place on every start, and one the
	// person can always write to, whatever folder the app was started from.
	notify := func(notification.Notice) {}
	if withWindow {
		notify = notification.Notify
	}
	binDir := filepath.Join(appDir, "bin")
	deps.EnsureIn(binDir, toolupdates.Reporter(logf, notify))
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return errors.New("'yt-dlp' could not be found or installed")
	}

	listener, existing, err := localserver.Listen(address)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", address, err)
	}
	if existing != "" {
		// Another MVD answers there. A window of this version finds that one's window and
		// raises it, then ends (see appwindow); an older version has no such window, so this
		// one opens its own on that server.
		fmt.Printf("MVD is already running at http://%s\n", existing)
		if withWindow {
			return appwindow.Run(context.Background(), appwindow.Options{URL: "http://" + existing, DataDir: appDir})
		}
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// yt-dlp and ffmpeg are kept up to date: checked in the background at start, and again
	// when a download fails, after which the failed downloads are tried again.
	var terminalRuns *terminalui.Runs
	updates := toolupdates.New(
		func() []string {
			return deps.UpdateIn(binDir, toolupdates.Reporter(logf, func(notification.Notice) {}))
		},
		func() int { return terminalRuns.RetryFailed() },
		notify, logf, time.Now,
	)
	terminalRuns = terminalui.NewRuns(updates.AfterFailure)

	removal := uninstall.Here(version, appDir, localserver.ConfigPath(appDir), stop)
	// The terminal-style interface: the same screens as the terminal app, in a window.
	terminal := terminalui.NewHandler(terminalui.NewModelFactory(appDir, terminalui.Files{
		Config: localserver.ConfigPath(appDir),
		List:   filepath.Join(appDir, "list.txt"),
	}, terminalui.Host{
		Start: terminalRuns.Start(ctx, ytDlpPath, logf),
		// The person is at this machine, so its own folder chooser is the one to show.
		PickFolder: func(start string) (string, bool, error) { return folderdialog.Dialog{}.Pick(ctx, start) },
		Uninstall:  terminalui.Uninstall(removal),
	}, logf), []string{"localhost:4200"})
	server := &http.Server{
		Handler:           localserver.NewHandler(removal, terminal),
		ReadHeaderTimeout: 10 * time.Second,
		// The terminal interface's open connections end when the app does, so shutting down is not held up.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
		stop()
	}()

	url := "http://" + listener.Addr().String()
	fmt.Printf("MVD %s at %s\n", version, url)
	updates.AtStart()

	var windowErr error
	if withWindow {
		// The window runs on this goroutine (macOS needs the main thread) until it is closed,
		// and closing it ends the app.
		windowErr = appwindow.Run(ctx, appwindow.Options{URL: url, DataDir: appDir, Busy: terminalRuns.Busy})
		stop()
	} else {
		fmt.Println("No window: Ctrl+C to quit.")
	}
	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		return err
	}
	if err := <-serveErr; err != nil {
		return err
	}

	return windowErr
}

// runUninstall removes the app. If one is already running it is asked to remove itself,
// which it does after asking the person and then quits; otherwise this process asks and
// removes. Either way nothing happens unless the person says yes.
func runUninstall(address string) error {
	if running, err := localserver.RequestUninstall(address); running {
		if errors.Is(err, uninstall.ErrDeclined) {
			return nil
		}

		return err
	}

	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}
	remove, err := uninstall.Here(version, appDir, localserver.ConfigPath(appDir), nil).Confirm(nil)
	switch {
	case errors.Is(err, uninstall.ErrDeclined):
		return nil
	case errors.Is(err, uninstall.ErrNoDialog):
		return errors.New("there is no way to ask for confirmation on this machine, so nothing was removed")
	case err != nil:
		return err
	}
	remove()

	return nil
}
