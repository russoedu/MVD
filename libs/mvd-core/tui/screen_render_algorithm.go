package tui

import (
	"fmt"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	helppage "youtube-downloader/libs/mvd-core/run/help"

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

	view := paddedFrame(downloadpage.View(m.downloadProps(m.width, m.height), m.width, m.height).View, m.width)

	if m.adding {
		return m.renderAddBox()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	return view
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
	props := helppage.Default()
	props.Palette = helppage.Palette(palette())
	width := 0
	for _, l := range lines {
		width = max(width, ansi.StringWidth(l))
	}
	props.HelpWidth, props.HelpHeight = width+4, len(lines)+2
	props.Lines = nil
	for _, l := range lines {
		props.Lines = append(props.Lines, helppage.LinesRow{Line: cellText(l)})
	}
	return paddedFrame(helppage.View(props, m.width, m.height).View, m.width)
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

// playlistGlyph is the mark at the end of a playlist row and its colour.
func (m model) playlistGlyph(pl *runstate.Playlist) (glyph, colour string) {
	c := currentColours
	if pl.Err != "" {
		return "✗", c.Error
	}
	if !pl.Listed {
		return spinnerFrames[m.spin], c.Dim
	}
	finished, failed, active := m.state.PlaylistTally(pl)
	switch {
	case active > 0:
		return spinnerFrames[m.spin], c.Highlight
	case finished == len(pl.Entries) && failed > 0:
		return "✗", c.Error
	case finished == len(pl.Entries):
		return "✓", c.Success
	}
	return "·", c.Dim
}

// entryGlyph is the mark at the start of an entry row and its colour.
func (m model) entryGlyph(en *runstate.Entry) (glyph, colour string) {
	c := currentColours
	switch en.State {
	case engine.StateQueued:
		return "·", c.Dim
	case engine.StateResolving:
		return "⟲", c.Focus
	case engine.StateDownloading:
		return spinnerFrames[m.spin], c.Highlight
	case engine.StateMerging:
		return "⚙", c.Highlight
	case engine.StateDone:
		return "✓", c.Success
	case engine.StateDuplicate:
		return "≡", c.Dim
	case engine.StateNotFound:
		return "∅", c.Dim
	case engine.StateFailed:
		return "✗", c.Error
	}
	return " ", c.Text
}

// entryTag is what is said at the end of an entry row and its colour.
func (m model) entryTag(en *runstate.Entry) (tag, colour string) {
	c := currentColours
	switch en.State {
	case engine.StateResolving:
		return "res.", c.Focus
	case engine.StateDownloading:
		if en.Total > 0 {
			return fmt.Sprintf("%3.0f%%", en.Percent), c.Highlight
		}
		return "dl", c.Highlight
	case engine.StateMerging:
		return "merge", c.Highlight
	case engine.StateDuplicate:
		return "dup", c.Dim
	case engine.StateNotFound:
		return "n/f", c.Dim
	case engine.StateFailed:
		return "ERR", c.Error
	case engine.StateDone:
		if en.Official {
			return "⇄", c.Focus
		}
		if en.Better {
			return "↑", c.Focus
		}
	}
	return "", c.Text
}
