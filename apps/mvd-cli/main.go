// MVD, Music Video Downloader. Settings and the download list live in the OS
// app-data folder (see internal/appdir). In a terminal it runs interactive
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

	"youtube-downloader/libs/mvd-core/appdir"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/deps"
	"youtube-downloader/libs/mvd-core/plain"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/sourcelist"
	"youtube-downloader/libs/mvd-core/tui"
)

func main() {
	noTUI := flag.Bool("no-tui", false, "plain log output instead of the interactive screens")
	flag.Parse()

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

	urls, _ := sourcelist.Load(listPath)
	if len(urls) == 0 {
		if legacy, _ := sourcelist.Load("downloads.conf"); len(legacy) > 0 {
			urls = legacy
			sourcelist.Save(listPath, urls)
		}
	}

	if tui.Enabled(*noTUI) {
		runInteractive(ytDlpPath, cfg, cfgPath, listPath, urls, created)
		return
	}
	runHeadless(ytDlpPath, cfg, cfgPath, listPath, urls, created)
}

// runInteractive loops between the setup screens and a download run.
func runInteractive(ytDlpPath string, cfg config.Config, cfgPath, listPath string, urls []string, created bool) {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	openConfig := created
	for {
		res, err := tui.RunSetup(tui.SetupInput{
			Cfg: cfg, URLs: urls, CfgPath: cfgPath, ListPath: listPath, OpenConfig: openConfig,
		})
		openConfig = false
		if err != nil {
			fmt.Printf("Interface error: %v\n", err)
			return
		}
		cfg, urls = res.Cfg, res.URLs
		if res.Action == tui.ActionQuit {
			return
		}
		if len(urls) == 0 {
			continue
		}
		if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
			fmt.Printf("Cannot create output directory %s: %v\n", cfg.OutputDir, err)
			continue
		}

		runCtx, cancel := context.WithCancel(appCtx)
		eng, err := runner.BuildEngine(runCtx, ytDlpPath, cfg, urls, func(f string, a ...interface{}) {
			fmt.Printf(f+"\n", a...)
		})
		if err != nil {
			cancel()
			fmt.Printf("Error: %v\n", err)
			continue
		}
		done := make(chan struct{})
		go func() { eng.Run(runCtx); close(done) }()

		state, _ := tui.RunDownload(runCtx, eng)
		cancel()
		for range eng.Events() {
		}
		<-done

		if state != nil {
			t := state.Tally()
			if t.Total > 0 && t.Queued == 0 && t.Running == 0 {
				sourcelist.Clear(listPath)
				urls = nil
			}
		}
		if appCtx.Err() != nil {
			return
		}
	}
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
		sourcelist.Clear(listPath)
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
	fmt.Printf("Official Music Video:     %v\n", cfg.DownloadOfficialMusicVideo)
	fmt.Printf("Auto Retry:               %v\n", cfg.AutoRetry)
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
