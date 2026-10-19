// MVD, Music Video Downloader. Settings and the download list live in the OS
// app-data folder (see libs/mvd-core/appdir). In a terminal it runs interactive
// screens (list, preferences, download); piped or with --no-tui it runs a
// plain headless download of the saved list.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/deps"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/openurl"
	"youtube-downloader/libs/mvd-core/plain"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/sourcelist"
	"youtube-downloader/libs/mvd-core/tui"
	"youtube-downloader/libs/mvd-core/workspace"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

func main() {
	noTUI := flag.Bool("no-tui", false, "plain log output instead of the interactive screens")
	edit := flag.String("edit", "", "open the playlist editor on a playlist address or a saved .mvd file; a .mvd file given by itself does the same")
	flag.Parse()

	review := *edit
	if review == "" && flag.NArg() == 1 && strings.EqualFold(filepath.Ext(flag.Arg(0)), playlistfile.Extension) {
		review = flag.Arg(0)
	}

	deps.Ensure()

	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("Error: 'yt-dlp' executable could not be found or installed.")
		os.Exit(1)
	}

	appDir, err := appdir.Dir()
	if err != nil {
		fmt.Printf("Error: cannot open app data directory: %v\n", err)
		os.Exit(1)
	}
	cfgPath := filepath.Join(appDir, "config.conf")
	listPath := filepath.Join(appDir, "list.txt")

	cfg, created, err := config.LoadOrCreate(cfgPath, appDir, appdir.DefaultDownloadsDir())
	if err != nil {
		fmt.Printf("Error reading config %s: %v\n", cfgPath, err)
		os.Exit(1)
	}

	tui.ApplyTheme(cfg.Colors)

	urls, _ := sourcelist.Load(listPath)
	if len(urls) == 0 {
		if legacy, _ := sourcelist.Load("downloads.conf"); len(legacy) > 0 {
			urls = legacy
			_ = sourcelist.Save(listPath, urls) // best effort: the import runs again next launch
		}
	}

	if tui.Enabled(*noTUI) {
		runInteractive(ytDlpPath, cfg, cfgPath, listPath, urls, created, appDir, review)
		return
	}
	runHeadless(ytDlpPath, cfg, cfgPath, listPath, urls, created)
}

// runInteractive runs the app model: the setup screens, the playlist editor and the
// download screen, one after the other, until the person quits.
func runInteractive(ytDlpPath string, cfg config.Config, cfgPath, listPath string, urls []string, created bool, appDir, review string) {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Nothing is printed while the screens are up; the engine writes its own log file.
	quiet := func(string, ...interface{}) {}
	startEngine := func(build func(context.Context) (*engine.Engine, error)) (*engineRun, error) {
		runCtx, cancel := context.WithCancel(appCtx)
		eng, err := build(runCtx)
		if err != nil {
			cancel()
			return nil, err
		}
		run := &engineRun{Engine: eng, cancel: cancel, done: make(chan struct{})}
		go func() { eng.Run(runCtx); close(run.done) }()
		return run, nil
	}

	list := downloadlist.New(urls)
	if review != "" {
		if strings.EqualFold(filepath.Ext(review), playlistfile.Extension) {
			opened, err := downloadlist.Open(review)
			if err != nil {
				fmt.Printf("Cannot open %s: %v\n", review, err)
				return
			}
			list = opened
		} else {
			list = downloadlist.New([]string{review})
		}
	}

	in := tui.WorkspaceInput{
		Setup: tui.SetupInput{Cfg: cfg, URLs: urls, CfgPath: cfgPath, ListPath: listPath, OpenConfig: created},
		List:  list,
		Live:  runner.LiveSettings,
		Host: workspace.Host{
			Start: func(cfg config.Config, sources []string, plan []engine.PlannedEntry, hold bool) (workspace.ListRun, error) {
				if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
					return nil, fmt.Errorf("cannot create the output directory %s: %w", cfg.OutputDir, err)
				}
				run, err := startEngine(func(ctx context.Context) (*engine.Engine, error) {
					return runner.BuildListEngine(ctx, ytDlpPath, cfg, sources, plan, hold, quiet)
				})
				if err != nil {
					return nil, err
				}
				return run, nil
			},
			Editor: tui.EditorHost{
				Open: openurl.Open,
				Describe: func(videoID string) (string, error) {
					return ytdlp.VideoTitle(appCtx, ytDlpPath, videoID, nil)
				},
				SessionFile: filepath.Join(appDir, "plan"+playlistfile.Extension),
				SaveDir:     filepath.Join(appDir, "plans"),
				DecisionLog: filepath.Join(appDir, "editor-decisions.jsonl"),
			},
			ListDir: filepath.Join(appDir, "lists"),
			Open:    openurl.Open,
		},
	}

	if err := tui.RunWorkspace(appCtx, in); err != nil && appCtx.Err() == nil {
		fmt.Printf("Interface error: %v\n", err)
	}
}

// engineRun is a run the app model can close: Close cancels it and waits until the engine
// has stopped and its events are drained.
type engineRun struct {
	*engine.Engine
	cancel context.CancelFunc
	done   chan struct{}
}

func (r *engineRun) Close() {
	r.cancel()
	for range r.Events() {
	}
	<-r.done
}

// runHeadless downloads the saved list without any screens.
func runHeadless(ytDlpPath string, cfg config.Config, cfgPath, listPath string, urls []string, created bool) {
	printBanner(cfg, cfgPath, listPath)
	if created {
		fmt.Printf("\nCreated a default config at %s\n", cfgPath)
	}
	if len(urls) == 0 {
		fmt.Printf("\nNo playlists or videos yet. Add them (one URL per line) to:\n  %s\nthen run MVD again.\n", listPath)
		return
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		fmt.Printf("Error creating output directory '%s': %v\n", cfg.OutputDir, err)
		os.Exit(1)
	}

	fmt.Printf("\nDownloading %d item(s), %d in parallel.\n\n", len(urls), cfg.MaxConcurrentDownloads)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	eng, err := runner.BuildEngine(ctx, ytDlpPath, cfg, urls, func(f string, a ...interface{}) {
		fmt.Printf(f+"\n", a...)
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() { eng.Run(ctx); close(done) }()

	tally := plain.Run(ctx, cancel, eng.Events(), eng.Sources(), os.Stdout)
	<-done

	if tally.Total > 0 && tally.Queued == 0 && tally.Running == 0 {
		if err := sourcelist.Clear(listPath); err != nil {
			fmt.Printf("Could not clear the saved list: %v\n", err)
		}
	}
	if tally.Failed > 0 {
		os.Exit(1)
	}
}

func printBanner(cfg config.Config, cfgPath, listPath string) {
	fmt.Println("\n--- MVD Configuration ---")
	fmt.Printf("Config file:              %s\n", cfgPath)
	fmt.Printf("List file:                %s\n", listPath)
	fmt.Printf("Output Directory:         %s\n", cfg.OutputDir)
	fmt.Printf("Video / Audio Quality:    %s / %s\n", cfg.VideoQuality, cfg.AudioQuality)
	fmt.Printf("Format (-f):              %s\n", cfg.Format())
	fmt.Printf("Merge Output Format:      %s\n", cfg.MergeOutputFormat)
	fmt.Printf("Output Template:          %s\n", cfg.OutputTemplate)
	fmt.Printf("Max Concurrent Downloads: %d\n", cfg.MaxConcurrentDownloads)
	fmt.Printf("Concurrent Fragments:     %d\n", cfg.ConcurrentFragments)
	fmt.Printf("Official Music Video:     %s\n", cfg.OfficialVideo)
	fmt.Printf("Auto Retry:               %v\n", cfg.AutoRetry)
	fmt.Printf("Retry Cycles / Wait:      %d / %ds x attempt\n", cfg.RetryCycles, cfg.RetryWaitSeconds)
	fmt.Printf("Cookies:                  %s\n", cookieDescription(cfg))
	fmt.Printf("Log File:                 %s\n", orNone(cfg.LogFile()))
	fmt.Println("-------------------------")
}

func cookieDescription(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser + " (pinned)"
	case cfg.AutoCookies:
		return "auto (all browsers)"
	default:
		return "off"
	}
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
