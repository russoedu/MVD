package tui

import (
	"fmt"
	"strings"
	"time"

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

	header := m.renderHeader(m.width)
	keys := m.renderKeyBar(m.width)
	bodyH := m.height - headerHeight - keyBarHeight
	if bodyH < 6 {
		bodyH = 6
	}

	var body string
	switch {
	case m.fullLog:
		body = m.renderOutput(m.width, bodyH)
	case m.width < minWideWidth:
		body = m.renderLeft(m.width, bodyH)
	default:
		leftW := m.width * 38 / 100
		if leftW < 36 {
			leftW = 36
		}
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderLeft(leftW, bodyH), m.renderOutput(m.width-leftW, bodyH))
	}

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, keys)

	if m.adding {
		return m.renderAddBox()
	}
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderHelp(), lipgloss.WithWhitespaceChars(" "))
	}
	return view
}

func (m model) renderHeader(w int) string {
	t := m.state.Tally()
	elapsed := runstate.HumanDuration(time.Since(m.state.Started))

	title := "MVD · Music Video Downloader"
	right := elapsed
	if m.state.Idle {
		right = "FINISHED · " + elapsed
	}

	parts := []string{
		styText.Render(fmt.Sprintf("Queue %d", t.Queued)),
		styYellow.Render(fmt.Sprintf("%s Running %d", spinnerFrames[m.spin], t.Running)),
		styGreen.Render(fmt.Sprintf("✓ Done %d", t.Done)),
		styCyan.Render(fmt.Sprintf("⇄ Official %d", t.Official)),
		styDim.Render(fmt.Sprintf("≡ Dup %d", t.Duplicate)),
		styRed.Render(fmt.Sprintf("✗ Failed %d", t.Failed)),
	}
	if t.Retried > 0 {
		parts = append(parts, styCyan.Render(fmt.Sprintf("↻ Retried %d", t.Retried)))
	}
	if m.state.Idle {
		parts[1] = styDim.Render(fmt.Sprintf("  Running %d", t.Running))
	}
	counters := strings.Join(parts, "   ")

	return renderBox(title, right, w, headerHeight, []string{counters}, false)
}

func (m model) renderLeft(w, h int) string {
	plH := len(m.state.Playlists) + 2
	maxPl := h / 3
	if maxPl < 5 {
		maxPl = 5
	}
	if plH > maxPl {
		plH = maxPl
	}
	if plH > h-5 {
		plH = h - 5
	}
	enH := h - plH
	return lipgloss.JoinVertical(lipgloss.Left, m.renderPlaylists(w, plH), m.renderEntries(w, enH))
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

func (m model) renderPlaylists(w, h int) string {
	innerW := w - 2
	rows := make([]string, 0, len(m.state.Playlists))
	for i, pl := range m.state.Playlists {
		finished, _, _ := m.state.PlaylistTally(pl)
		count := fmt.Sprintf("%d/%d", finished, len(pl.Entries))
		if pl.Err != "" {
			count = "error"
		} else if !pl.Listed {
			count = "listing"
		}
		cursor := " "
		if i == m.selPlaylist {
			cursor = "▸"
		}
		titleW := innerW - 2 - 1 - len(count) - 1 - 2
		title := padRight(truncate(display(pl.Title), titleW), titleW)
		row := fmt.Sprintf("%s %s %s ", cursor, title, count)
		if i == m.selPlaylist {
			row = stySelected.Render(row)
		} else {
			row = styText.Render(row)
		}
		rows = append(rows, row+m.playlistGlyph(pl))
	}
	rows = window(rows, m.selPlaylist, h-2)
	return renderBox("Playlists", fmt.Sprintf("%d", len(m.state.Playlists)), w, h, rows, m.focus == paneLists)
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
	case engine.StateFailed:
		return styRed.Render("ERR"), 3
	case engine.StateDone:
		if en.Official {
			return styCyan.Render("⇄"), 1
		}
	}
	return "", 0
}

func (m model) renderEntries(w, h int) string {
	innerW := w - 2
	ids := m.visibleEntries()
	sel := clamp(m.selEntry, 0, max(0, len(ids)-1))

	rows := make([]string, 0, len(ids))
	for i, id := range ids {
		en := m.state.Entries[id]
		tag, tagW := m.entryTag(en)
		// glyph(1) space(1) idx(2) space(1) title space(1) tag
		titleW := innerW - 1 - 1 - 2 - 1 - 1 - tagW
		title := padRight(truncate(entryTitle(en), titleW), titleW)
		text := fmt.Sprintf(" %02d %s ", en.Index, title)
		if i == sel && m.focus == paneEntries {
			text = stySelected.Render(text)
		} else if i == sel {
			text = styMagenta.Render(text)
		} else {
			text = styText.Render(text)
		}
		rows = append(rows, m.entryGlyph(en)+text+tag)
	}
	rows = window(rows, sel, h-2)

	title := "Entries"
	if m.filterFailed {
		title = "Entries (failed only)"
	}
	right := ""
	if m.selPlaylist < len(m.state.Playlists) {
		pl := m.state.Playlists[m.selPlaylist]
		if pl.Err != "" {
			rows = nil
			for _, l := range wrap("✗ "+display(pl.Err), innerW) {
				rows = append(rows, styRed.Render(l))
			}
		} else if !pl.Listed {
			rows = []string{styDim.Render("listing playlist...")}
		} else if len(ids) == 0 && m.filterFailed {
			rows = []string{styDim.Render("no failed entries")}
		}
		right = fmt.Sprintf("%d", len(ids))
	}
	return renderBox(title, right, w, h, rows, m.focus == paneEntries)
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

func (m model) renderOutput(w, h int) string {
	innerW := w - 2
	innerH := h - 2
	lines := m.outputLines()

	var footer []string
	if en := m.selectedEntry(); en != nil && m.focus != paneLists {
		switch en.State {
		case engine.StateDownloading, engine.StateMerging:
			footer = append(footer, "", m.renderBar(en, innerW))
		case engine.StateFailed:
			footer = append(footer, "", styRed.Render(truncate("✗ "+display(en.Err), innerW)))
		case engine.StateDone:
			what := "original " + en.TargetID
			if en.Official {
				what = "official video " + en.TargetID
			}
			footer = append(footer, "", styGreen.Render(truncate("✓ downloaded "+what, innerW)))
		}
	}

	avail := innerH - len(footer)
	if avail < 1 {
		avail = 1
	}
	end := len(lines) - m.outScroll
	if end < 0 {
		end = 0
	}
	start := end - avail
	if start < 0 {
		start = 0
	}
	shown := make([]string, 0, avail+len(footer))
	for _, l := range lines[start:end] {
		shown = append(shown, styText.Render(truncate(display(l), innerW)))
	}
	for len(shown) < avail {
		shown = append(shown, "")
	}
	shown = append(shown, footer...)

	right := ""
	if m.follow {
		right = "follow"
	}
	if m.outScroll > 0 {
		right = fmt.Sprintf("↑%d", m.outScroll)
	}
	return renderBox(m.outputTitle(), right, w, h, shown, m.focus == paneOutput)
}

func (m model) renderBar(en *runstate.Entry, w int) string {
	info := runstate.ProgressLine(en)
	if en.State == engine.StateMerging {
		info = "merging streams with ffmpeg"
	}
	barW := w - ansi.StringWidth(info) - 2
	if barW < 10 {
		return styYellow.Render(truncate(info, w))
	}
	filled := int(float64(barW) * en.Percent / 100)
	if en.State == engine.StateMerging || en.Total == 0 {
		filled = barW
	}
	filled = clamp(filled, 0, barW)
	bar := styBarFill.Render(strings.Repeat("█", filled)) + styBarEmpty.Render(strings.Repeat("░", barW-filled))
	return bar + "  " + styYellow.Render(info)
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

func (m model) renderKeyBar(w int) string {
	if m.confirmQuit {
		return fit(styRed.Render(" Downloads are still running. Quit? ")+styKey.Render("y")+styDim.Render("/")+styKey.Render("n"), w)
	}
	if m.status != "" && time.Now().Before(m.statusUntil) {
		return fit(" "+styMagenta.Render(m.status), w)
	}
	keys := m.hints()
	var b strings.Builder
	b.WriteString(" ")
	for i, k := range keys {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(styKey.Render(k.key) + " " + styDim.Render(k.desc))
	}
	if m.state.Idle {
		b.WriteString("   " + styGreen.Render("all done"))
	}
	return fit(b.String(), w)
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
