// MVD desktop app: a window of its own that shows the terminal interface (the same
// screens as the terminal app), carried over the window's own events with no server or
// port, and that takes new URLs while it downloads. Closing the window quits. The terminal app is apps/mvd-tui; both share
// libs/mvd-core.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"youtube-downloader/apps/mvd/appwindow"
	"youtube-downloader/apps/mvd/console"
	"youtube-downloader/apps/mvd/folderdialog"
	"youtube-downloader/apps/mvd/install"
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

	// Older shortcuts and scripts pass these; there is no server to point or to run without a
	// window any more, so they are accepted and do nothing.
	flag.String("addr", "", "no longer used (kept for older scripts)")
	flag.Bool("no-window", false, "no longer used (kept for older scripts)")
	flag.Bool("no-browser", false, "no longer used (kept for older scripts)")
	flag.Bool("no-tray", false, "no longer used (kept for older scripts)")
	move := flag.Bool("move", false, "ask to move the app to its own folder now, whatever was answered before")
	movedFrom := flag.String("moved-from", "", "set by the app itself after moving to its folder: the old copy to remove")
	removeApp := flag.Bool("uninstall", false, "remove MVD from this computer, after asking; Settings > Apps on Windows runs this")
	flag.Parse()

	if *removeApp {
		if err := runUninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			console.ShowFatal(err.Error())
			os.Exit(1)
		}

		return
	}

	if err := run(*movedFrom, *move); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		console.ShowFatal(err.Error())
		os.Exit(1)
	}
}

func run(movedFrom string, askToMove bool) error {
	logf := func(format string, a ...interface{}) { fmt.Printf(format+"\n", a...) }

	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}

	// The first time it is started from somewhere it does not belong, it offers to move
	// itself, and if that is accepted the moved copy takes over and this one is done.
	// -move asks whatever was answered before.
	if askToMove {
		switch install.MoveNow(appDir, version) {
		case install.Moved:
			return nil
		case install.AlreadyThere:
			logf("MVD already lives in its own folder.")
		case install.CannotMove:
			logf("MVD cannot be moved on this machine.")
		}
	} else if install.OfferMoveHere(appDir, true, movedFrom, version) {
		return nil
	}
	if movedFrom != "" {
		go install.CleanUpMovedProgram(movedFrom)
	}

	// The tools live in the app-data folder: the same place on every start, and one the
	// person can always write to, whatever folder the app was started from.
	notify := notification.Notify
	binDir := filepath.Join(appDir, "bin")
	deps.EnsureIn(binDir, toolupdates.Reporter(logf, notify))
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return errors.New("'yt-dlp' could not be found or installed")
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

	removal := uninstall.Here(version, appDir, terminalui.ConfigPath(appDir), stop)
	// The terminal-style interface: the same screens as the terminal app, in a window.
	program := terminalui.NewShared(terminalui.NewModelFactory(appDir, terminalui.Files{
		Config: terminalui.ConfigPath(appDir),
		List:   filepath.Join(appDir, "list.txt"),
	}, terminalui.Host{
		Start: terminalRuns.Start(ctx, ytDlpPath, logf),
		// The person is at this machine, so its own folder chooser is the one to show.
		PickFolder: func(start string) (string, bool, error) { return folderdialog.Dialog{}.Pick(ctx, start) },
		Uninstall:  terminalui.Uninstall(removal),
		// Moving later is possible whatever was answered when the app asked by itself.
		Move: terminalui.Move(func() install.MoveResult { return install.MoveNow(appDir, version) }, stop),
	}, logf))

	fmt.Printf("MVD %s\n", version)
	updates.AtStart()

	// The window runs on this goroutine (macOS needs the main thread) until it is closed,
	// and closing it ends the app. A second start does not get past appwindow.Run: it
	// raises the first window and ends.
	err = appwindow.Run(ctx, appwindow.Options{Program: program, DataDir: appDir, Busy: terminalRuns.Busy})
	stop()

	return err
}

// runUninstall removes the app: this process asks the person and removes it, and nothing
// happens unless they say yes. A copy that is running keeps running until it is closed;
// removing the program moves it aside, which Windows allows while it runs.
func runUninstall() error {
	appDir, err := appdir.Dir()
	if err != nil {
		return fmt.Errorf("cannot open the app data folder: %w", err)
	}
	remove, err := uninstall.Here(version, appDir, terminalui.ConfigPath(appDir), nil).Confirm(nil)
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
