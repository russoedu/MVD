package download

import (
	"fmt"
	"time"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/run/loading"
	helppage "youtube-downloader/libs/mvd-core/run/help"
	"youtube-downloader/libs/mvd-core/screenview"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return "starting..."
	}
	if m.width < 60 || m.height < 16 {
		return look.Dim.Render(fmt.Sprintf("terminal too small (%dx%d), need at least 60x16", m.width, m.height))
	}

	if !m.session.Loaded && !m.session.Hosted && !m.session.Adding && !m.session.ShowHelp {
		return m.renderLoading()
	}

	view := screenview.Padded(View(m.downloadProps(m.width, m.height), m.width, m.height).View, m.width)
	view = m.withTableControls(m.withChoiceButtons(view))

	if m.session.Adding {
		return m.renderAddBox()
	}
	if m.session.ShowHelp {
		return m.renderHelp()
	}
	return view
}

// renderLoading is the screen shown until every playlist of the run has been listed.
func (m Model) renderLoading() string {
	p := loading.Progress{Spin: m.spin, Detail: "Reading the songs of each playlist", Elapsed: runstate.HumanDuration(time.Since(m.session.State.Started))}
	for _, pl := range m.session.State.Playlists {
		p.Playlists = append(p.Playlists, loading.Playlist{
			Name: m.playlistName(pl), Songs: len(pl.Entries), Listed: pl.Listed, Failed: pl.Err != "",
		})
	}
	if m.session.ConfirmQuit {
		p.Confirm = look.Red.Render(" Downloads are still running. Quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	}
	return loading.Render(p, m.width, m.height)
}

func (m Model) outputTitle() string {
	if !m.session.Technical || m.session.SelPlaylist >= len(m.session.State.Playlists) {
		return "Details"
	}
	pl := m.session.State.Playlists[m.session.SelPlaylist]
	if m.session.Pane == paneLists || m.session.SelectedEntry() == nil {
		return "Technical log · " + cells.Display(m.playlistName(pl))
	}
	en := m.session.SelectedEntry()
	return fmt.Sprintf("Technical log · %02d %s", en.Index, entryTitle(en))
}

// hints lists the keys the download screen offers now. The key bar and the
// outline both read it, so they cannot drift apart.
func (m Model) hints() []keybar.Hint {
	var hints []keybar.Hint
	if m.session.State.Idle {
		hints = []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "tab", Desc: "pane"}, {Key: "f", Desc: "failed"}, {Key: "r", Desc: "retry failed"}, {Key: "t", Desc: "technical log"}}
	} else {
		hints = []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "tab", Desc: "pane"}, {Key: "⏎", Desc: "follow"}, {Key: "f", Desc: "failed"}, {Key: "r", Desc: "retry"}, {Key: "t", Desc: "technical log"}}
	}
	if m.session.CanChoose() && (m.session.Pane == paneEntries || len(m.session.Picked) > 0) && len(m.session.TargetIDs()) > 0 {
		hints = append([]keybar.Hint{{Key: "space", Desc: "select"}, {Key: "d", Desc: "accept"}, {Key: "x", Desc: "reject"}, {Key: "o", Desc: "open video"}, {Key: "D", Desc: "accept all"}}, append(hints[:2:2], keybar.Hint{Key: "t", Desc: "technical log"})...)
	}
	if m.session.Technical {
		// The log of the tool is not the screen people expect: the way back is always on the bar.
		kept := hints[:0:0]
		for _, h := range hints {
			if h.Key != "t" {
				kept = append(kept, h)
			}
		}
		hints = append(kept, keybar.Hint{Key: "t", Desc: "back to the details"})
	}
	if m.session.Hosted {
		return hints
	}
	if m.session.CanAdd() {
		hints = append(hints, keybar.Hint{Key: "a", Desc: "add links"})
	}
	if !m.session.State.Idle {
		hints = append(hints, keybar.Hint{Key: "?", Desc: "help"})
	}
	return append(hints, keybar.Hint{Key: "q", Desc: "quit"})
}

func (m Model) renderHelp() string {
	logPath := ""
	if m.session.Ctrl != nil {
		logPath = m.session.Ctrl.LogPath()
	}
	return helppage.Screen(m.width, m.height, helppage.Options{LogPath: logPath, Hosted: m.session.Hosted})
}

// downloadingSummary says how many songs are coming in now and how fast, all of them together:
// "Downloading 4 songs · 12.4MiB/s". The speed is left out while it is not known.
func downloadingSummary(t runstate.Tally) string {
	summary := "Downloading " + counted(t.Downloading, "song")
	if t.Speed > 0 {
		summary += " · " + runstate.HumanSpeed(t.Speed)
	}
	return summary
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
	finished, failed, active := m.session.State.PlaylistTally(pl)
	switch {
	case active > 0:
		return look.SpinnerFrames[m.spin], c.Highlight
	case m.session.WaitingIn(pl) > 0:
		return "?", c.Highlight
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
	case engine.StatePlanned:
		if songlist.FoundOfficial(en) {
			return "✓", c.Success
		}

		return "?", c.Highlight
	case engine.StateAccepted:
		return "▸", c.Focus
	case engine.StateSkipped:
		return "–", c.Dim
	}
	return " ", c.Text
}

