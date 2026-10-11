package tui

import (
	"fmt"
	"youtube-downloader/libs/mvd-core/screens"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func (m model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	if m.width < 60 || m.height < 16 {
		return styDim.Render(fmt.Sprintf("terminal too small (%dx%d), need at least 60x16", m.width, m.height))
	}

	view := paddedFrame(screens.Download(m.downloadProps(m.width, m.height), m.width, m.height).View, m.width)

	if m.adding {
		return m.renderAddBox()
	}
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderHelp(), lipgloss.WithWhitespaceChars(" "))
	}
	return view
}

func (m model) playlistGlyph(pl *runstate.Playlist) string {
	if pl.Err != "" {
		return styRed.Render("✗")
	}
	if !pl.Listed {
		return styDim.Render(spinnerFrames[m.spin])
	}
	finished, failed, active := m.state.PlaylistTally(pl)
	switch {
	case active > 0:
		return styYellow.Render(spinnerFrames[m.spin])
	case finished == len(pl.Entries) && failed > 0:
		return styRed.Render("✗")
	case finished == len(pl.Entries):
		return styGreen.Render("✓")
	}
	return styDim.Render("·")
}

func (m model) entryGlyph(en *runstate.Entry) string {
	switch en.State {
	case engine.StateQueued:
		return styDim.Render("·")
	case engine.StateResolving:
		return styCyan.Render("⟲")
	case engine.StateDownloading:
		return styYellow.Render(spinnerFrames[m.spin])
	case engine.StateMerging:
		return styYellow.Render("⚙")
	case engine.StateDone:
		return styGreen.Render("✓")
	case engine.StateDuplicate:
		return styDim.Render("≡")
	case engine.StateNotFound:
		return styDim.Render("∅")
	case engine.StateFailed:
		return styRed.Render("✗")
	}
	return " "
}

func (m model) entryTag(en *runstate.Entry) (string, int) {
	switch en.State {
	case engine.StateResolving:
		return styCyan.Render("res."), 4
	case engine.StateDownloading:
		if en.Total > 0 {
			s := fmt.Sprintf("%3.0f%%", en.Percent)
			return styYellow.Render(s), len(s)
		}
		return styYellow.Render("dl"), 2
	case engine.StateMerging:
		return styYellow.Render("merge"), 5
	case engine.StateDuplicate:
		return styDim.Render("dup"), 3
	case engine.StateNotFound:
		return styDim.Render("n/f"), 3
	case engine.StateFailed:
		return styRed.Render("ERR"), 3
	case engine.StateDone:
		if en.Official {
			return styCyan.Render("⇄"), 1
		}
		if en.Better {
			return styCyan.Render("↑"), 1
		}
	}
	return "", 0
}

// outputLines returns the log lines for the current selection.
func (m model) outputLines() []string {
	if m.focus == paneLists || m.selectedEntry() == nil {
		if m.selPlaylist < len(m.state.Playlists) {
			return m.state.Playlists[m.selPlaylist].Log
		}
		return nil
	}
	return m.selectedEntry().Log
}

func (m model) outputTitle() string {
	if m.selPlaylist >= len(m.state.Playlists) {
		return "Output"
	}
	pl := m.state.Playlists[m.selPlaylist]
	if m.focus == paneLists || m.selectedEntry() == nil {
		return display(pl.Title)
	}
	en := m.selectedEntry()
	return fmt.Sprintf("%s › %02d %s", display(pl.Title), en.Index, entryTitle(en))
}

// hints lists the keys the download screen offers now. The key bar and the
// outline both read it, so they cannot drift apart.
func (m model) hints() []keyHint {
	var hints []keyHint
	if m.state.Idle {
		hints = []keyHint{{"↑↓", "move"}, {"tab", "pane"}, {"f", "failed"}, {"r", "retry failed"}, {"l", "full log"}}
	} else {
		hints = []keyHint{{"↑↓", "move"}, {"tab", "pane"}, {"⏎", "follow"}, {"f", "failed"}, {"r", "retry"}, {"l", "full log"}}
	}
	if m.canAdd() {
		hints = append(hints, keyHint{"a", "add links"})
	}
	if !m.state.Idle {
		hints = append(hints, keyHint{"?", "help"})
	}
	return append(hints, keyHint{"q", "quit"})
}

func (m model) renderHelp() string {
	maxW := m.width - 6
	if maxW < 20 {
		maxW = 20
	}
	var lines []string
	for _, l := range []string{
		styKey.Render("↑ ↓ j k") + "     move selection, scroll output",
		styKey.Render("pgup pgdn") + "   move ten rows",
		styKey.Render("tab") + "         cycle panes: playlists, entries, output",
		styKey.Render("enter") + "       toggle follow mode (jump to the active download)",
		styKey.Render("f") + "           show only failed entries",
		styKey.Render("r") + "           retry the selected failed entry, or all failed in the playlist",
		styKey.Render("l") + "           toggle full screen output",
		styKey.Render("a") + "           add more links to this run",
		styKey.Render("q") + "           quit (asks for confirmation while running)",
		styKey.Render("ctrl+c") + "      quit immediately",
		styKey.Render("ctrl+l") + "      redraw the screen",
		"",
		styDim.Render("glyphs: · queued  ⟲ resolving  ⠋ downloading  ⚙ merging  ✓ done  ≡ duplicate  ✗ failed  ⇄ official video"),
		"",
		styDim.Render("full output is written to " + m.ctrl.LogPath()),
		"",
		styDim.Render("press any key to close"),
	} {
		if ansi.StringWidth(l) <= maxW {
			lines = append(lines, l)
			continue
		}
		lines = append(lines, wrap(l, maxW)...)
	}
	if len(lines) > m.height-2 {
		lines = lines[:m.height-2]
	}
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return renderBox("Help", "", w+4, len(lines)+2, lines, true)
}

// downloadingSummary is the speed of all the downloads together, and how many files and
// how many parts of them are coming in now: "↓ 12.4MiB/s · 4 files · 9 parts".
func downloadingSummary(t runstate.Tally) string {
	return fmt.Sprintf("↓ %s · %s · %s", runstate.HumanSpeed(t.Speed), counted(t.Downloading, "file"), counted(t.Parts, "part"))
}

// counted writes a count with its noun: "1 file", "4 files".
func counted(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
