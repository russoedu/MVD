// MVD tray app: runs until it is closed, serves a browser UI on localhost, and
// takes new URLs while it downloads. The terminal app is apps/mvd-cli; both share
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

	"youtube-downloader/apps/mvd-tray/browser"
	"youtube-downloader/apps/mvd-tray/console"
	"youtube-downloader/apps/mvd-tray/folderdialog"
	"youtube-downloader/apps/mvd-tray/install"
	"youtube-downloader/apps/mvd-tray/localserver"
	"youtube-downloader/apps/mvd-tray/notification"
	"youtube-downloader/apps/mvd-tray/toolupdates"
	"youtube-downloader/apps/mvd-tray/tray"
	"youtube-downloader/apps/mvd-tray/uninstall"
	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/deps"
	"youtube-downloader/libs/mvd-server/api"
	"youtube-downloader/libs/mvd-server/session"
	"youtube-downloader/libs/mvd-server/settings"
)

var version = "dev"

func main() {
	console.AttachParent()

	address := flag.String("addr", localserver.DefaultAddress, "address to serve the UI on; keep it on 127.0.0.1")
	noBrowser := flag.Bool("no-browser", false, "do not open the UI in the browser on start")
	noTray := flag.Bool("no-tray", false, "do not put an icon in the system tray (run until Ctrl+C)")
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

	if err := run(*address, !*noBrowser, !*noTray, *movedFrom); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		console.ShowFatal(err.Error())
		os.Exit(1)
	}
}

func run(address string, open, withTray bool, movedFrom string) error {
	logf := func(format string, a ...interface{}) { fmt.Printf(format+"\n", a...) }

	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}

	// The first time it is started from somewhere it does not belong, it offers to move
	// itself, and if that is accepted the moved copy takes over and this one is done.
	if install.OfferMoveHere(appDir, withTray, movedFrom, version) {
		return nil
	}
	if movedFrom != "" {
		go install.CleanUpMovedProgram(movedFrom)
	}

	// The tools live in the app-data folder: the same place on every start, and one the
	// person can always write to, whatever folder the app was started from.
	notify := func(notification.Notice) {}
	if withTray {
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
		fmt.Printf("MVD is already running at http://%s\n", existing)
		if open {
			_ = browser.Open("http://" + existing)
		}
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// yt-dlp and ffmpeg are kept up to date: checked in the background at start, and again
	// when a download fails, after which the failed downloads are tried again.
	var sessions *session.Session
	updates := toolupdates.New(
		func() []string {
			return deps.UpdateIn(binDir, toolupdates.Reporter(logf, func(notification.Notice) {}))
		},
		func() int { return sessions.RetryFailed() },
		notify, logf, time.Now,
	)
	sessions = session.New(ctx, localserver.NewEngineFactory(ytDlpPath, appDir, logf), session.WithFailureHook(updates.AfterFailure))
	defer sessions.Close()

	removal := uninstall.Here(version, appDir, localserver.ConfigPath(appDir), stop)
	server := &http.Server{
		Handler:           localserver.NewHandler(sessions, settings.NewRepository(localserver.ConfigPath(appDir), appDir, appdir.DefaultDownloadsDir()), folderdialog.Dialog{}, removal),
		ReadHeaderTimeout: 10 * time.Second,
		// Open event streams end when the app does, so shutting down is not held up.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	url := "http://" + listener.Addr().String()
	quitHint := "Ctrl+C to quit"
	if withTray {
		quitHint = "use the tray icon or Ctrl+C to quit"
	}
	fmt.Printf("MVD %s at %s (%s)\n", version, url, quitHint)
	updates.AtStart()
	if open {
		if err := browser.Open(url); err != nil {
			fmt.Printf("Open %s in your browser.\n", url)
		}
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

	if withTray {
		tray.Run(ctx, url, browser.Open, stop)
		if ctx.Err() == nil {
			fmt.Println("No system tray is available here; running without an icon (Ctrl+C to quit).")
			tray.Unavailable(url)
		}
	}
	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		return err
	}

	return <-serveErr
}

// runUninstall removes the app. If one is already running it is asked to remove itself,
// which it does after asking the person and then quits; otherwise this process asks and
// removes. Either way nothing happens unless the person says yes.
func runUninstall(address string) error {
	if running, err := localserver.RequestUninstall(address); running {
		if errors.Is(err, api.ErrUninstallDeclined) {
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
	case errors.Is(err, api.ErrUninstallDeclined):
		return nil
	case errors.Is(err, api.ErrNoDialog):
		return errors.New("there is no way to ask for confirmation on this machine, so nothing was removed")
	case err != nil:
		return err
	}
	remove()

	return nil
}
