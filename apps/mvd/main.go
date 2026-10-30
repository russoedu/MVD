// MVD desktop app: a window of its own that shows the terminal interface (the same
// screens as the terminal app), carried over the window's own events with no server or
// port, and that takes new URLs while it downloads. Closing the window quits, and so does quitting from its screens. The terminal app is apps/mvd-tui; both share
// libs/mvd-core.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"youtube-downloader/apps/mvd/appwindow"
	"youtube-downloader/apps/mvd/console"
	"youtube-downloader/apps/mvd/filetype"
	"youtube-downloader/apps/mvd/folderdialog"
	"youtube-downloader/apps/mvd/install"
	"youtube-downloader/apps/mvd/notification"
	"youtube-downloader/apps/mvd/openfolder"
	"youtube-downloader/apps/mvd/question"
	"youtube-downloader/apps/mvd/selfupdate"
	"youtube-downloader/apps/mvd/signin"
	"youtube-downloader/apps/mvd/terminalui"
	"youtube-downloader/apps/mvd/toolupdates"
	"youtube-downloader/apps/mvd/uninstall"
	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/cookies"
	"youtube-downloader/libs/mvd-core/deps"
	"youtube-downloader/libs/mvd-core/openurl"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/prefsform"
	"youtube-downloader/libs/mvd-core/prefspopup"
	"youtube-downloader/libs/mvd-core/tui"
	"youtube-downloader/libs/mvd-core/workspace"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
	"youtube-downloader/libs/mvd-core/ytdlp"
	"youtube-downloader/libs/mvd-core/ytdlp/fetch"
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
	edit := flag.String("edit", "", "open the playlist editor on a playlist address or a saved .mvd file; a .mvd file given by itself does the same")
	afterExit := flag.Int(selfupdate.AfterExitFlag, 0, "set by the app itself after an update: wait for this process to end before opening")
	removeApp := flag.Bool("uninstall", false, "remove MVD from this computer, after asking; Settings > Apps on Windows runs this")
	signInDir := flag.String(signin.Flag, "", "set by the app itself: show the window to sign in to YouTube, with its profile in this folder")
	flag.Parse()

	if *signInDir != "" {
		if err := signin.RunWindow(*signInDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	// An updated program is started by the old one, which is still closing: it opens only once
	// that one has gone, so the two never share the single window.
	if *afterExit > 0 {
		selfupdate.WaitForExit(*afterExit, 30*time.Second)
	}
	// The old program an update renamed aside is removed, once nothing runs it.
	if exe, err := os.Executable(); err == nil {
		go install.CleanUpMovedProgram(exe + selfupdate.AsideSuffix)
	}

	if *removeApp {
		if err := runUninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			console.ShowFatal(err.Error())
			os.Exit(1)
		}

		return
	}

	review := *edit
	if review == "" && flag.NArg() == 1 && strings.EqualFold(filepath.Ext(flag.Arg(0)), playlistfile.Extension) {
		review = flag.Arg(0)
	}
	if err := run(*movedFrom, *move, review); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		console.ShowFatal(err.Error())
		os.Exit(1)
	}
}

func run(movedFrom string, askToMove bool, review string) error {
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
	// Fetched in the background on the first start, so the window opens at once and says what is
	// going on; the downloads wait for it (see Host.Tools).
	tools := deps.PrepareIn(binDir, toolupdates.Reporter(logf, notify))
	ytDlpPath := deps.ToolPath(binDir, "yt-dlp")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// yt-dlp and ffmpeg are kept up to date: checked in the background at start, and again
	// when a download fails, after which the failed downloads are tried again.
	var terminalRuns *terminalui.Runs
	checker := &selfupdate.Checker{
		Dir: appDir, Version: version, Notify: notify, Ask: question.Ask, Open: openurl.Open, Logf: logf,
		// The switch is in the preferences, read each time so a change counts at once.
		Enabled: func() bool {
			cfg, err := config.Load(terminalui.ConfigPath(appDir), appDir)
			return err != nil || cfg.CheckForUpdates
		},
	}
	updates := toolupdates.New(
		func() []string {
			return deps.UpdateIn(binDir, toolupdates.Reporter(logf, func(notification.Notice) {}))
		},
		func() int { return terminalRuns.RetryFailed() },
		notify, logf, time.Now,
	)
	terminalRuns = terminalui.NewRuns(updates.AfterFailure)

	// The page around the terminal dims with a popup, and is told when one is open.
	backdrop := &appwindow.Backdrop{}
	// The window can be drawn in the classic systems' looks: the program says what it shows,
	// and the page sends back what the person does.
	classic := &appwindow.ClassicWindow{}
	removal := uninstall.Here(version, appDir, terminalui.ConfigPath(appDir), stop)
	// Moving is only offered while the app has not been moved yet.
	var move tui.Mover
	if install.CanBeMoved(appDir, version) {
		move = terminalui.Move(func() install.MoveResult { return install.MoveNow(appDir, version) }, stop)
	}
	// The terminal-style interface: the same screens as the terminal app, in a window.
	program := terminalui.NewShared(terminalui.NewModelFactory(appDir, terminalui.Files{
		Config: terminalui.ConfigPath(appDir),
		List:   filepath.Join(appDir, "list.txt"),
		Review: review,
	}, terminalui.Host{
		Start:        terminalRuns.Start(ctx, ytDlpPath, logf),
		StartList:    terminalRuns.StartList(ctx, ytDlpPath, logf),
		Plan:         terminalRuns.StartPlan(ctx, ytDlpPath, logf),
		PlanDownload: terminalRuns.StartPlanDownload(ctx, ytDlpPath, logf),
		Describe: func(videoID string) (string, error) {
			return ytdlp.VideoTitle(ctx, ytDlpPath, videoID, nil)
		},
		// The person is at this machine, so its own folder chooser is the one to show.
		PickFolder: func(start string) (string, bool, error) { return folderdialog.Dialog{}.Pick(ctx, start) },
		PickFile: func(start string) (string, bool, error) {
			return folderdialog.Dialog{}.PickFile(ctx, start, "mvd", "MVD review")
		},
		SaveFile: func(dir, name string) (string, bool, error) {
			return folderdialog.Dialog{}.SaveFile(ctx, dir, name, "mvd", "MVD review")
		},
		// Quitting from the screens ends the program that runs them, and the window would
		// otherwise stay open on a page with nothing behind it.
		Quit:      stop,
		Version:   version,
		Clipboard: terminalui.SystemClipboard{},

		CheckForUpdate: func() string { return checker.CheckNow(ctx) },
		UpdateNote:     checker.Note,
		Modal:          backdrop.Set,
		Style:          classic.SetStyle,
		Window:         classic.Show,
		// The preferences are a window the page draws, in every style.
		Preferences: classic.ShowPreferences,
		Tools:       toolsState(tools),
		OpenFolder:  openfolder.Open,
		Uninstall:   terminalui.Uninstall(removal),
		// Moving later is possible whatever was answered when the app asked by itself.
		Move: move,
		// Signing in inside MVD is only done on Windows, where yt-dlp can read the window's cookies.
		SignIn:      signInHost(ctx, appDir, ytDlpPath),
		SignInState: func() string { return cookies.SignInState(appDir) },
	}, logf))

	// .mvd files are download lists of this program: they get its icon and open it. It is
	// told at every start, so it follows the program when it moves.
	if exe, err := os.Executable(); err == nil {
		if err := filetype.Register(exe); err != nil {
			logf("cannot register the .mvd file type: %v", err)
		}
	}

	fmt.Printf("MVD %s\n", version)
	// Looking for newer tools waits for the first fetch, so the two never write the same file.
	go func() {
		_ = tools.Wait(ctx)
		updates.AtStart()
	}()
	// A newer MVD is looked for once a day, in the background; the person is only told.
	go checker.AtStart(ctx)

	// The window runs on this goroutine (macOS needs the main thread) until it is closed,
	// and closing it ends the app. A second start does not get past appwindow.Run: it
	// raises the first window and ends.
	classic.OnIntent = func(in windowsnapshot.Intent) { program.Send(workspace.WindowIntentMsg(in)) }
	classic.OnPrefsIntent = func(in prefsform.WindowIntent) { program.Send(prefspopup.WindowIntentMsg(in)) }
	checker.Quit = func() { program.Send(workspace.CloseRequestedMsg{}) }
	err = appwindow.Run(ctx, appwindow.Options{
		Program: program, DataDir: appDir, Backdrop: backdrop, Classic: classic,
		OnCloseRequest: func() { program.Send(workspace.CloseRequestedMsg{}) },
		// A list double clicked while the app is open is opened in the window that is.
		OnSecondStart: func(args []string) {
			if list := filetype.ListInArgs(args); list != "" {
				program.Send(workspace.OpenListMsg{Path: list})
			}
		},
		// macOS hands a double clicked .mvd over this way instead of in the arguments.
		OnFileOpen: func(path string) {
			if list := filetype.ListInArgs([]string{path}); list != "" {
				program.Send(workspace.OpenListMsg{Path: list})
			}
		},
	})
	stop()
	// An update put in place while the app ran opens as soon as the app has closed.
	if exe := checker.Relaunch(); exe != "" {
		if startErr := selfupdate.Relaunch(exe); startErr != nil {
			logf("The updated MVD could not be started: %v", startErr)
		}
	}

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

// toolsState is what the workspace asks about the tools: ready, being fetched, or one that
// could not be had.
func toolsState(tools *deps.Preparation) func() workspace.ToolsState {
	return func() workspace.ToolsState {
		if problem := tools.Problem(); problem != "" {
			return workspace.ToolsState{Message: problem, Failed: true}
		}
		if tools.Ready() {
			return workspace.ToolsState{Ready: true}
		}

		return workspace.ToolsState{Message: tools.Status()}
	}
}

// signInHost is what the preferences call to sign in to YouTube inside MVD, or nil where that is
// not offered.
func signInHost(ctx context.Context, appDir, ytDlpPath string) func() (string, error) {
	if runtime.GOOS != "windows" {
		return nil
	}
	ytArgs := fetch.NormalizeArgs(config.Default(appDir, "").ExtraArgs)

	return func() (string, error) { return signin.Run(ctx, appDir, ytDlpPath, ytArgs) }
}
