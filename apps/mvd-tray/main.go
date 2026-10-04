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
)

var version = "dev"

func main() {
	address := flag.String("addr", defaultAddress, "address to serve the UI on; keep it on 127.0.0.1")
	noBrowser := flag.Bool("no-browser", false, "do not open the UI in the browser on start")
	flag.Parse()

	if err := run(*address, !*noBrowser); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(address string, open bool) error {
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
		Handler:           newAppHandler(sessions),
		ReadHeaderTimeout: 10 * time.Second,
		// Open event streams end when the app does, so shutting down is not held up.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	url := "http://" + listener.Addr().String()
	fmt.Printf("MVD %s at %s (Ctrl+C to quit)\n", version, url)
	if open {
		if err := openBrowser(url); err != nil {
			fmt.Printf("Open %s in your browser.\n", url)
		}
	}

	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()

	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdown)
}
