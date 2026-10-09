package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/ytdlp"
)

const (
	exitChanged = 1
	exitError   = 2
)

func main() {
	playlist := flag.String("playlist", os.Getenv("YT_CHECK_PLAYLIST"), "the playlist kept for the check (or $YT_CHECK_PLAYLIST)")
	baselinePath := flag.String("baseline", "tools/youtube-check/baseline.json", "the recorded answers")
	record := flag.Bool("record", false, "record the answers of now as the baseline instead of comparing")
	maxDifferent := flag.Float64("max-different", 0.1, "the share of the songs that may answer differently before it counts as a change")
	reportPath := flag.String("report", "", "write the report (Markdown) to this file as well")
	ytDlp := flag.String("yt-dlp", "", "the yt-dlp program (default: the one on the PATH)")
	workers := flag.Int("workers", 3, "songs looked up at the same time")
	flag.Parse()

	code, err := run(*playlist, *baselinePath, *record, *maxDifferent, *reportPath, *ytDlp, *workers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "youtube-check: %v\n", err)
		os.Exit(exitError)
	}
	os.Exit(code)
}

func run(playlist, baselinePath string, record bool, maxDifferent float64, reportPath, ytDlp string, workers int) (int, error) {
	if playlist == "" {
		return 0, fmt.Errorf("no playlist: pass -playlist or set YT_CHECK_PLAYLIST")
	}
	if ytDlp == "" {
		path, err := exec.LookPath("yt-dlp")
		if err != nil {
			return 0, fmt.Errorf("yt-dlp is not on the PATH")
		}
		ytDlp = path
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	baseline, err := loadBaseline(baselinePath)
	if err != nil {
		return 0, err
	}
	entries, err := ytdlp.ListPlaylist(ctx, ytDlp, playlist, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot read the playlist: %w", err)
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("the playlist has no songs")
	}

	// No cache: it would answer from earlier runs and hide what YouTube does now.
	resolver, _ := runner.BuildResolver(ctx, runner.ResolverInput{YtDlp: ytDlp}, func(format string, a ...interface{}) {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
	})
	lookup := songLookup{
		Wanted: func(title string) bool { return resolver.Wanted(title, "", "") },
		Find: func(entry ytdlp.PlaylistEntry) (string, string) {
			channel := entry.Channel
			if channel == "" {
				channel = entry.Uploader
			}
			return resolver.ResolveLog(entry.ID, entry.Title, channel, int(entry.Duration), nil)
		},
	}
	fmt.Fprintf(os.Stderr, "looking up %d songs...\n", len(entries))
	outcomes := checkSongs(entries, baseline, lookup, workers)

	if record {
		if err := saveBaseline(baselinePath, baselineFrom(playlist, outcomes)); err != nil {
			return 0, err
		}
		fmt.Printf("Recorded the answers for %d songs in %s\n", len(outcomes), baselinePath)
		return 0, nil
	}

	verdict := Judge(baseline, outcomes, maxDifferent)
	report := reportOf(playlist, verdict, maxDifferent)
	fmt.Println(report)
	if reportPath != "" {
		if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
			return 0, err
		}
	}
	if verdict.Failed {
		return exitChanged, nil
	}

	return 0, nil
}
