package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// useTUI decides between the full screen interface and plain log output.
func useTUI(noTUIFlag bool) bool {
	if noTUIFlag || os.Getenv("MVD_NO_TUI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// runPlain prints engine events as log lines. It is used when stdout is
// not a terminal (CI, pipes) or when --no-tui is given. It returns once
// the engine reports idle.
func runPlain(ctx context.Context, cancel context.CancelFunc, eng *Engine, out io.Writer) Tally {
	state := newRunState(eng.Sources())
	lastStep := map[int]int{}

	tag := func(en *entryView) string {
		return fmt.Sprintf("[P%d/%02d]", en.Playlist+1, en.Index)
	}

	for ev := range eng.Events() {
		id := state.apply(ev)
		switch e := ev.(type) {
		case EvPlaylistListed:
			fmt.Fprintf(out, "[P%d] %q: %d entries\n", e.Playlist+1, e.Title, len(e.Entries))
		case EvPlaylistFailed:
			fmt.Fprintf(out, "[P%d] FAILED to list: %s\n", e.Playlist+1, e.Err)
		case EvEntryState:
			en := state.entry(id)
			switch e.State {
			case StateResolving:
				fmt.Fprintf(out, "%s resolving official video for %s\n", tag(en), en.Title)
			case StateDownloading:
				what := "original"
				if en.Official {
					what = "official video " + en.TargetID
				}
				fmt.Fprintf(out, "%s downloading %s (%s)\n", tag(en), en.Title, what)
				lastStep[id] = -1
			case StateMerging:
				fmt.Fprintf(out, "%s merging\n", tag(en))
			case StateDone:
				fmt.Fprintf(out, "%s DONE %s\n", tag(en), en.Title)
			case StateDuplicate:
				fmt.Fprintf(out, "%s duplicate of an earlier entry, skipped\n", tag(en))
			case StateFailed:
				fmt.Fprintf(out, "%s FAILED %s: %s\n", tag(en), en.Title, en.Err)
			}
		case EvProgress:
			en := state.entry(id)
			step := int(en.Percent) / 10
			if en.Total > 0 && step > lastStep[id] {
				lastStep[id] = step
				fmt.Fprintf(out, "%s %s\n", tag(en), progressLine(en))
			}
		case EvLog:
			if e.Entry < 0 {
				fmt.Fprintf(out, "[P%d] %s\n", e.Playlist+1, e.Line)
			} else if strings.HasPrefix(e.Line, "ERROR") || strings.HasPrefix(e.Line, "[official]") {
				fmt.Fprintf(out, "%s %s\n", tag(state.entry(e.Entry)), e.Line)
			}
		case EvIdle:
			cancel()
		}
	}

	t := state.tally()
	printSummary(out, state, t)
	return t
}

func printSummary(out io.Writer, state *runState, t Tally) {
	fmt.Fprintln(out, "==================================================")
	fmt.Fprintln(out, "Download Summary:")
	fmt.Fprintf(out, "Playlists:              %d\n", len(state.Playlists))
	fmt.Fprintf(out, "Videos total:           %d\n", t.Total)
	fmt.Fprintf(out, "  downloaded:           %d\n", t.Done)
	fmt.Fprintf(out, "  replaced by official: %d\n", t.Official)
	fmt.Fprintf(out, "  skipped duplicates:   %d\n", t.Duplicate)
	fmt.Fprintf(out, "  failed:               %d\n", t.Failed)
	for _, pl := range state.Playlists {
		if pl.Err != "" {
			fmt.Fprintf(out, "  [P%d] %s: %s\n", pl.Index+1, pl.URL, pl.Err)
		}
	}
	for _, en := range state.Entries {
		if en != nil && en.State == StateFailed {
			fmt.Fprintf(out, "  [P%d/%02d] %s: %s\n", en.Playlist+1, en.Index, en.Title, en.Err)
		}
	}
	fmt.Fprintln(out, "==================================================")
}
