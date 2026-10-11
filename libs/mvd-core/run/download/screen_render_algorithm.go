package download

import (
	"fmt"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	helppage "youtube-downloader/libs/mvd-core/run/help"
	"youtube-downloader/libs/mvd-core/screenview"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	if m.width < 60 || m.height < 16 {
		return look.Dim.Render(fmt.Sprintf("terminal too small (%dx%d), need at least 60x16", m.width, m.height))
	}

	view := screenview.Padded(View(m.downloadProps(m.width, m.height), m.width, m.height).View, m.width)

	if m.adding {
		return m.renderAddBox()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	return view
}

// outputLines returns the log lines for the current selection.
func (m Model) outputLines() []string {
	if m.focus == paneLists || m.selectedEntry() == nil {
		if m.selPlaylist < len(m.state.Playlists) {
			return m.state.Playlists[m.selPlaylist].Log
		}
		return nil
	}
	return m.selectedEntry().Log
}

func (m Model) outputTitle() string {
	if m.selPlaylist >= len(m.state.Playlists) {
		return "Output"
	}
	pl := m.state.Playlists[m.selPlaylist]
	if m.focus == paneLists || m.selectedEntry() == nil {
		return cells.Display(pl.Title)
	}
	en := m.selectedEntry()
	return fmt.Sprintf("%s › %02d %s", cells.Display(pl.Title), en.Index, entryTitle(en))
}

// hints lists the keys the download screen offers now. The key bar and the
// outline both read it, so they cannot drift apart.
func (m Model) hints() []keybar.Hint {
	var hints []keybar.Hint
	if m.state.Idle {
		hints = []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "tab", Desc: "pane"}, {Key: "f", Desc: "failed"}, {Key: "r", Desc: "retry failed"}, {Key: "l", Desc: "full log"}}
	} else {
		hints = []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "tab", Desc: "pane"}, {Key: "⏎", Desc: "follow"}, {Key: "f", Desc: "failed"}, {Key: "r", Desc: "retry"}, {Key: "l", Desc: "full log"}}
	}
	if m.canAdd() {
		hints = append(hints, keybar.Hint{Key: "a", Desc: "add links"})
	}
	if !m.state.Idle {
		hints = append(hints, keybar.Hint{Key: "?", Desc: "help"})
	}
	return append(hints, keybar.Hint{Key: "q", Desc: "quit"})
}

func (m Model) renderHelp() string {
	return helppage.Screen(m.width, m.height, m.ctrl.LogPath())
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
func (m Model) playlistGlyph(pl *runstate.Playlist) (glyph, colour string) {
	c := look.Colours
	if pl.Err != "" {
		return "✗", c.Error
	}
	if !pl.Listed {
		return look.SpinnerFrames[m.spin], c.Dim
	}
	finished, failed, active := m.state.PlaylistTally(pl)
	switch {
	case active > 0:
		return look.SpinnerFrames[m.spin], c.Highlight
	case finished == len(pl.Entries) && failed > 0:
		return "✗", c.Error
	case finished == len(pl.Entries):
		return "✓", c.Success
	}
	return "·", c.Dim
}

// entryGlyph is the mark at the start of an entry row and its colour.
func (m Model) entryGlyph(en *runstate.Entry) (glyph, colour string) {
	c := look.Colours
	switch en.State {
	case engine.StateQueued:
		return "·", c.Dim
	case engine.StateResolving:
		return "⟲", c.Focus
	case engine.StateDownloading:
		return look.SpinnerFrames[m.spin], c.Highlight
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
func (m Model) entryTag(en *runstate.Entry) (tag, colour string) {
	c := look.Colours
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
