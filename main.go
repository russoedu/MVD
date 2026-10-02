// MVD, Music Video Downloader: reads playlist URLs from downloads.conf and
// settings from setup.conf, makes sure yt-dlp and ffmpeg are available and
// downloads every playlist, optionally swapping auto-generated art tracks
// for the official music video. This file only wires the slices together;
// each slice under internal/ owns one outcome.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"youtube-downloader/internal/config"
	"youtube-downloader/internal/deps"
	"youtube-downloader/internal/engine"
	"youtube-downloader/internal/official"
	"youtube-downloader/internal/plain"
	"youtube-downloader/internal/runstate"
	"youtube-downloader/internal/tui"
)

func main() {
	noTUI := flag.Bool("no-tui", false, "plain log output instead of the full screen interface")
	flag.Parse()

	// 1. Auto-check & download any missing dependencies
	deps.Ensure()

	setupFile := "setup.conf"
	downloadsFile := "downloads.conf"

	// 2. Verify yt-dlp location
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Println("Error: 'yt-dlp' executable could not be found or installed.")
		os.Exit(1)
	}
	fmt.Printf("Found yt-dlp at: %s\n", ytDlpPath)

	// 3. Load settings from setup.conf
	cfg, err := config.LoadSetup(setupFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", setupFile, err)
		os.Exit(1)
	}

	fmt.Println("\n--- Downloader Configuration ---")
	fmt.Printf("Output Directory:         %s\n", cfg.OutputDir)
	fmt.Printf("Quality:                  %s\n", cfg.Quality)
	fmt.Printf("Merge Output Format:      %s\n", cfg.MergeOutputFormat)
	fmt.Printf("Output Template:          %s\n", cfg.OutputTemplate)
	fmt.Printf("Max Concurrent Downloads: %d\n", cfg.MaxConcurrentDownloads)
	fmt.Printf("Extra Arguments:          %v\n", cfg.ExtraArgs)
	fmt.Printf("Official Music Video:     %v\n", cfg.DownloadOfficialMusicVideo)
	fmt.Printf("Log File:                 %s\n", cfg.LogFile)
	fmt.Println("--------------------------------")

	// 4. Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory '%s': %v\n", cfg.OutputDir, err)
		os.Exit(1)
	}

	// 5. Load URLs from downloads.conf
	urls, err := config.LoadDownloads(downloadsFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", downloadsFile, err)
		os.Exit(1)
	}

	if len(urls) == 0 {
		fmt.Printf("No valid playlist URLs found in '%s'.\n", downloadsFile)
		return
	}

	fmt.Printf("Found %d playlist(s) to process (%d parallel download(s)).\n\n", len(urls), cfg.MaxConcurrentDownloads)

	// 6. Build the engine and attach a renderer
	opts := engine.Options{
		YtDlp:             ytDlpPath,
		URLs:              urls,
		OutputDir:         cfg.OutputDir,
		OutputTemplate:    cfg.OutputTemplate,
		Quality:           cfg.Quality,
		MergeOutputFormat: cfg.MergeOutputFormat,
		ExtraArgs:         cfg.ExtraArgs,
		Workers:           cfg.MaxConcurrentDownloads,
		LogPath:           cfg.LogFile,
	}
	if cfg.DownloadOfficialMusicVideo {
		opts.Resolver = official.NewResolver(nil)
	}

	eng, err := engine.New(opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	engineDone := make(chan struct{})
	go func() {
		eng.Run(ctx)
		close(engineDone)
	}()

	var tally runstate.Tally
	if tui.Enabled(*noTUI) {
		state, err := tui.Run(ctx, eng)
		cancel()
		// Unblock the engine so it can shut down, then print the summary
		// into the normal screen buffer where it stays in the scrollback.
		for range eng.Events() {
		}
		<-engineDone
		if err != nil {
			fmt.Printf("Interface error: %v\n", err)
		}
		if state != nil {
			tally = state.Tally()
			plain.PrintSummary(os.Stdout, state, tally)
		}
	} else {
		tally = plain.Run(ctx, cancel, eng.Events(), eng.Sources(), os.Stdout)
		<-engineDone
	}

	if tally.Failed > 0 {
		os.Exit(1)
	}
}
