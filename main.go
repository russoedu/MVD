// MVD, Music Video Downloader. Settings and the download list live in the OS
// app-data folder (see internal/appdir); this file wires the slices together:
// load config, build the engine from it, and drive a renderer.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"youtube-downloader/internal/appdir"
	"youtube-downloader/internal/config"
	"youtube-downloader/internal/deps"
	"youtube-downloader/internal/plain"
	"youtube-downloader/internal/runner"
	"youtube-downloader/internal/runstate"
	"youtube-downloader/internal/sourcelist"
	"youtube-downloader/internal/tui"
)

func main() {
	noTUI := flag.Bool("no-tui", false, "plain log output instead of the full screen interface")
	flag.Parse()

	// 1. Auto-check & download any missing dependencies.
	deps.Ensure()

	// 2. Verify yt-dlp location.
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("Error: 'yt-dlp' executable could not be found or installed.")
		os.Exit(1)
	}
	fmt.Printf("Found yt-dlp at: %s\n", ytDlpPath)

	// 3. Locate app-data and load (or create) the config.
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

	// 4. Load the saved list, importing a legacy ./downloads.conf once.
	urls, _ := sourcelist.Load(listPath)
	if len(urls) == 0 {
		if legacy, _ := sourcelist.Load("downloads.conf"); len(legacy) > 0 {
			urls = legacy
			sourcelist.Save(listPath, urls)
		}
	}

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

	engineDone := make(chan struct{})
	go func() {
		eng.Run(ctx)
		close(engineDone)
	}()

	var tally runstate.Tally
	if tui.Enabled(*noTUI) {
		state, rerr := tui.Run(ctx, eng)
		cancel()
		for range eng.Events() {
		}
		<-engineDone
		if rerr != nil {
			fmt.Printf("Interface error: %v\n", rerr)
		}
		if state != nil {
			tally = state.Tally()
			plain.PrintSummary(os.Stdout, state, tally)
		}
	} else {
		tally = plain.Run(ctx, cancel, eng.Events(), eng.Sources(), os.Stdout)
		<-engineDone
	}

	// When the whole list was processed, clear it so the next run starts fresh.
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
