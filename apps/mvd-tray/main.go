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
	"time"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/deps"
	"youtube-downloader/libs/mvd-server/session"
	"youtube-downloader/libs/mvd-server/settings"
)

var version = "dev"

func main() {
	attachParentConsole()

	address := flag.String("addr", defaultAddress, "address to serve the UI on; keep it on 127.0.0.1")
	noBrowser := flag.Bool("no-browser", false, "do not open the UI in the browser on start")
	noTray := flag.Bool("no-tray", false, "do not put an icon in the system tray (run until Ctrl+C)")
	flag.Parse()

	if err := run(*address, !*noBrowser, !*noTray); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		showFatal(err.Error())
		os.Exit(1)
	}
}

func run(address string, open, tray bool) error {
	deps.Ensure()
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return errors.New("'yt-dlp' could not be found or installed")
	}
	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}

	listener, existing, err := listen(address)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", address, err)
	}
	if existing != "" {
		fmt.Printf("MVD is already running at http://%s\n", existing)
		if open {
			_ = openBrowser("http://" + existing)
		}
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logf := func(format string, a ...interface{}) { fmt.Printf(format+"\n", a...) }
	sessions := session.New(ctx, newEngineFactory(ytDlpPath, appDir, logf))
	defer sessions.Close()

	server := &http.Server{
		Handler:           newAppHandler(sessions, settings.NewRepository(configPath(appDir), appDir, appdir.DefaultDownloadsDir()), folderDialog{}),
		ReadHeaderTimeout: 10 * time.Second,
		// Open event streams end when the app does, so shutting down is not held up.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	url := "http://" + listener.Addr().String()
	quitHint := "Ctrl+C to quit"
	if tray {
		quitHint = "use the tray icon or Ctrl+C to quit"
	}
	fmt.Printf("MVD %s at %s (%s)\n", version, url, quitHint)
	if open {
		if err := openBrowser(url); err != nil {
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

	if tray {
		runTray(ctx, url, openBrowser, stop)
		if ctx.Err() == nil {
			fmt.Println("No system tray is available here; running without an icon (Ctrl+C to quit).")
			trayUnavailable(url)
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
