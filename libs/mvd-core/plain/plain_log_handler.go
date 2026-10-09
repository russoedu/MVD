// Package plain prints the run as log lines, for pipes, CI and --no-tui.
package plain

import (
	"context"
	"fmt"
	"io"
	"strings"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Run prints engine events as log lines until the engine reports idle,
// then cancels the context so the engine shuts down, and keeps draining
// until the events channel closes. It returns the final counters.
func Run(ctx context.Context, cancel context.CancelFunc, events <-chan interface{}, sources []engine.PlaylistSource, out io.Writer) runstate.Tally {
	state := runstate.New(sources)
	lastStep := map[int]int{}

	tag := func(en *runstate.Entry) string {
		return fmt.Sprintf("[P%d/%02d]", en.Playlist+1, en.Index)
	}

	for ev := range events {
		id := state.Apply(ev)
		switch e := ev.(type) {
		case engine.EvPlaylistListing:
			_, _ = fmt.Fprintf(out, "[P%d] listing %s\n", e.Playlist+1, e.URL)
		case engine.EvPlaylistListed:
			_, _ = fmt.Fprintf(out, "[P%d] %q: %d entries\n", e.Playlist+1, e.Title, len(e.Entries))
		case engine.EvPlaylistFailed:
			_, _ = fmt.Fprintf(out, "[P%d] FAILED to list: %s\n", e.Playlist+1, e.Err)
		case engine.EvEntryState:
			en := state.Entry(id)
			if en == nil {
				continue
			}
			switch e.State {
			case engine.StateResolving:
				_, _ = fmt.Fprintf(out, "%s resolving official video for %s\n", tag(en), en.Title)
			case engine.StateDownloading:
				what := "original"
				if en.Official {
					what = "official video " + en.TargetID
				} else if en.Better {
					what = "better quality video " + en.TargetID
				}
				_, _ = fmt.Fprintf(out, "%s downloading %s (%s)\n", tag(en), en.Title, what)
				lastStep[id] = -1
			case engine.StateMerging:
				_, _ = fmt.Fprintf(out, "%s merging\n", tag(en))
			case engine.StateDone:
				_, _ = fmt.Fprintf(out, "%s DONE %s\n", tag(en), en.Title)
			case engine.StateDuplicate:
				_, _ = fmt.Fprintf(out, "%s duplicate of an earlier entry, skipped\n", tag(en))
			case engine.StateFailed:
				_, _ = fmt.Fprintf(out, "%s FAILED %s: %s\n", tag(en), en.Title, en.Err)
			}
		case engine.EvProgress:
			en := state.Entry(id)
			if en == nil {
				continue
			}
			step := int(en.Percent) / 10
			if en.Total > 0 && step > lastStep[id] {
				lastStep[id] = step
				_, _ = fmt.Fprintf(out, "%s %s\n", tag(en), runstate.ProgressLine(en))
			}
		case engine.EvLog:
			// Playlist level lines duplicate the listing events above.
			if en := state.Entry(e.Entry); en != nil && (strings.HasPrefix(e.Line, "ERROR") || strings.HasPrefix(e.Line, "[official]") || strings.HasPrefix(e.Line, "official video ")) {
				_, _ = fmt.Fprintf(out, "%s %s\n", tag(en), e.Line)
			}
		case engine.EvIdle:
			cancel()
		}
	}

	t := state.Tally()
	PrintSummary(out, state, t)
	return t
}

// PrintSummary writes the end of run report.
func PrintSummary(out io.Writer, state *runstate.State, t runstate.Tally) {
	_, _ = fmt.Fprintln(out, "==================================================")
	_, _ = fmt.Fprintln(out, "Download Summary:")
	_, _ = fmt.Fprintf(out, "Playlists:              %d\n", len(state.Playlists))
	_, _ = fmt.Fprintf(out, "Videos total:           %d\n", t.Total)
	_, _ = fmt.Fprintf(out, "  downloaded:           %d\n", t.Done)
	_, _ = fmt.Fprintf(out, "  replaced by official: %d\n", t.Official)
	_, _ = fmt.Fprintf(out, "  replaced by better quality: %d\n", t.Better)
	_, _ = fmt.Fprintf(out, "  skipped duplicates:   %d\n", t.Duplicate)
	_, _ = fmt.Fprintf(out, "  failed:               %d\n", t.Failed)
	_, _ = fmt.Fprintf(out, "Retried:                %d\n", t.Retried)
	for _, pl := range state.Playlists {
		if pl.Err != "" {
			_, _ = fmt.Fprintf(out, "  [P%d] %s: %s\n", pl.Index+1, pl.URL, pl.Err)
		}
	}
	for _, en := range state.Entries {
		if en != nil && en.State == engine.StateFailed {
			_, _ = fmt.Fprintf(out, "  [P%d/%02d] %s: %s\n", en.Playlist+1, en.Index, en.Title, en.Err)
		}
	}
	_, _ = fmt.Fprintln(out, "==================================================")
}
