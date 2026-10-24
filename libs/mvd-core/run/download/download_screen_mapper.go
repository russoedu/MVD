package download

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/keybar"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/internal/cuppa"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/viewport"
)

// downloadProps fills the properties of the Download screen, drawn from its design, for
// a window of w by h cells. The design owns where everything goes; what is here is what
// the screen shows. The rows come as functions of the room the design gives them, so the
// screen cuts them to fit and a click is matched with the page's own At functions.
func (m Model) downloadProps(w, h int) Props {
	p := Default()
	p.Palette = Palette(look.CurrentPalette())

	// The sizes the design cannot know: how wide the left column is.
	p.LeftWidth = m.leftWidth(w)
	if m.columns(w) {
		p.Columns, p.MidWidth, p.LeftReserve = 1, m.midWidth(w), m.reserve
	}
	p.FullLog = 0
	if m.fullLog {
		p.FullLog = 1
	}
	showOutput := m.fullLog || w >= minWideWidth
	p.PlaylistCount = len(m.state.Playlists)

	p.Counters, p.HeaderRight = m.headerTexts()
	p.Counters = cells.Truncate(p.Counters, w-2)

	// The app has a footer under the output only when there is something for it.
	footer := m.outputFooter()
	if !showOutput {
		footer = outputFooter{}
	}
	if len(footer.lines) > 0 {
		p.FooterLines = 2
	}

	// Playlists.
	p.PlaylistsRows = func(w, h int) ([]PlaylistsRow, int) {
		return cuppa.Window(m.playlistRows(w), m.selPlaylist, h)
	}
	p.PlaylistsRight = fmt.Sprintf("%d", len(m.state.Playlists))
	p.PlaylistsFocus = m.focus == paneLists

	// Entries.
	p.EntriesTitle, p.EntriesRight = m.entryHeader()
	p.EntriesRows = func(w, h int) ([]EntriesRow, int) {
		rows, sel := m.entryRows(w)
		return cuppa.Window(rows, sel, h)
	}
	p.HasNotes = 0
	if m.entryNotesShown() {
		p.HasNotes = 1
	}
	p.NotesRows = func(w, _ int) ([]NotesRow, int) { return m.entryNotes(w), 0 }
	p.EntriesFocus = m.focus == paneEntries

	// Output.
	p.OutputRows = func(w, h int) ([]OutputRow, int) {
		lines, first := m.outputWindow(w, h)
		rows := make([]OutputRow, 0, len(lines))
		for _, line := range lines {
			rows = append(rows, OutputRow{Line: cells.CellText(line)})
		}
		return rows, first
	}
	p.OutputTitle = m.outputTitle()
	p.OutputRight = ""
	if m.follow {
		p.OutputRight = "follow"
	}
	if m.outScroll > 0 {
		p.OutputRight = fmt.Sprintf("↑%d", m.outScroll)
	}
	p.OutputFocus = m.focus == paneOutput

	// The footer of the output pane: a progress bar, or a note.
	if len(footer.lines) > 0 {
		outputBox, _ := Layout(p, w, h).RegionNamed("Output box")
		inner := outputBox.W - 2
		switch {
		case footer.bar != nil && inner-ansi.StringWidth(footer.bar.info)-2 >= 10:
			p.ShowProgress = true
			p.BarWidth = inner - ansi.StringWidth(footer.bar.info) - 2
			p.Percent = footer.bar.percent
			p.ProgressText = footer.bar.info
		default:
			p.ShowNote = true
			p.Note = cells.Truncate(footer.lines[len(footer.lines)-1], inner)
		}
	}

	// The bottom line: what was asked, what just happened, or the keys.
	p.Hints = keybar.Of[HintsHint](m.hints())
	switch {
	case m.confirmQuit:
		p.ConfirmQuit = true
		p.Confirm = look.Red.Render(" Downloads are still running. Quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	case m.status != "" && time.Now().Before(m.statusUntil):
		p.ShowStatus = true
		p.Status = " " + m.status
	default:
		// The keys are for a run; while the list is only being made there is nothing to press.
		p.ShowKeys = !m.editable
	}
	return p
}

// headerTexts are the counters inside the header box and the label on its top edge.
func (m Model) headerTexts() (counters, right string) {
	t := m.state.Tally()
	elapsed := runstate.HumanDuration(time.Since(m.state.Started))

	right = elapsed
	if m.editable {
		right = "" // nothing is running: there is no time to count
	} else if m.state.Idle {
		right = "FINISHED · " + elapsed
	} else if t.Downloading > 0 {
		right = downloadingSummary(t) + " · " + elapsed
	}

	parts := []string{
		look.Text.Render(fmt.Sprintf("Queue %d", t.Queued)),
		look.Yellow.Render(fmt.Sprintf("%s Running %d", look.SpinnerFrames[m.spin], t.Running)),
		look.Green.Render(fmt.Sprintf("✓ Done %d", t.Done)),
		look.Cyan.Render(fmt.Sprintf("⇄ Official %d", t.Official)),
		look.Cyan.Render(fmt.Sprintf("↑ Better %d", t.Better)),
		look.Dim.Render(fmt.Sprintf("≡ Dup %d", t.Duplicate)),
		look.Red.Render(fmt.Sprintf("✗ Failed %d", t.Failed)),
	}
	if t.NotFound > 0 {
		parts = append(parts, look.Dim.Render(fmt.Sprintf("∅ Not found %d", t.NotFound)))
	}
	if t.Retried > 0 {
		parts = append(parts, look.Cyan.Render(fmt.Sprintf("↻ Retried %d", t.Retried)))
	}
	if m.state.Idle {
		parts[1] = look.Dim.Render(fmt.Sprintf("  Running %d", t.Running))
	}
	return strings.Join(parts, "   "), right
}

// playlistRows are the rows of the playlists pane: the design places the cursor, the
// title, the count and the glyph; the app says what each says and the colour of the glyph.
func (m Model) playlistRows(w int) []PlaylistsRow {
	if len(m.state.Playlists) == 0 {
		var rows []PlaylistsRow
		for _, line := range wrapText("Add music: playlists or videos, with the Add link button or by pasting links", w-6) {
			rows = append(rows, PlaylistsRow{Mark: " ", Sp: " ", Title: line, TitleStyle: CellStyle{Fg: look.Colours.Dim}})
		}
		return rows
	}
	rows := make([]PlaylistsRow, 0, len(m.state.Playlists))
	for i, pl := range m.state.Playlists {
		finished, _, _ := m.state.PlaylistTally(pl)
		count := fmt.Sprintf("%d/%d", finished, len(pl.Entries))
		if pl.Err != "" {
			count = "error"
		} else if !pl.Listed {
			count = "listing"
		}
		glyph, colour := m.playlistGlyph(pl)
		title := cells.CellText(cells.Display(pl.Title))
		row := PlaylistsRow{
			Mark: " ", Title: title, Count: count, Glyph: glyph,
			GlyphStyle: CellStyle{Fg: colour},
		}
		if m.editable {
			row.Mark, row.Count, row.Glyph = "≡", "✎", "×"
			row.GlyphStyle, row.CountStyle = CellStyle{Fg: look.Colours.Error}, CellStyle{Fg: look.Colours.Focus}
		}
		if m.dragRowOn && i == m.dropRow {
			row.Mark = "▶"
		}
		if i == m.selPlaylist {
			// 6 cells are the mark, the gaps, the glyph and the end of the row.
			row.Style = RowSelected
			if !m.editable {
				row.Mark = "▸"
			}
			row.Title = marquee(title, w-6-ansi.StringWidth(row.Count), m.marquee)
		}
		rows = append(rows, row)
	}
	return rows
}

// entryRows are the rows of the entries pane for the selected playlist and the row that
// is selected.
func (m Model) entryRows(w int) (rows []EntriesRow, sel int) {
	ids := m.visibleEntries()
	if len(ids) == 0 {
		var rows []EntriesRow
		for _, line := range wrapText("The songs of the selected playlist show here", w-6) {
			rows = append(rows, EntriesRow{Glyph: " ", Sp: " ", Title: line, TitleStyle: CellStyle{Fg: look.Colours.Dim}})
		}
		return rows, 0
	}
	sel = viewport.Clamp(m.selEntry, 0, max(0, len(ids)-1))

	for i, id := range ids {
		en := m.state.Entries[id]
		glyph, glyphColour := m.entryGlyph(en)
		tag, tagColour := m.entryTag(en)
		row := EntriesRow{
			Glyph: glyph, Idx: fmt.Sprintf("%02d", en.Index), Title: cells.CellText(entryTitle(en)), Tag: tag,
			GlyphStyle: CellStyle{Fg: glyphColour}, TagStyle: CellStyle{Fg: tagColour},
		}
		if i == sel {
			// Whether the pane has the focus is the design's: an unfocused pane draws it as text.
			row.Style = RowSelected
			// 6 cells are the glyph, the number, the gaps and the end of the row.
			row.Title = marquee(row.Title, w-6-ansi.StringWidth(tag), m.marquee)
		}
		rows = append(rows, row)
	}
	return rows, sel
}

// entryHeader are the title and the label of the entries pane.
func (m Model) entryHeader() (title, right string) {
	title = "Entries"
	if m.filterFailed {
		title = "Entries (failed only)"
	}
	if m.selPlaylist < len(m.state.Playlists) {
		right = fmt.Sprintf("%d", len(m.visibleEntries()))
	}
	return title, right
}

// entryNotesShown is whether the entries pane has a note to show instead of rows: the
// playlist has an error, is still being listed, or has no failed entries to show.
func (m Model) entryNotesShown() bool {
	if m.selPlaylist >= len(m.state.Playlists) {
		return false
	}
	pl := m.state.Playlists[m.selPlaylist]
	return pl.Err != "" || !pl.Listed || (m.filterFailed && len(m.visibleEntries()) == 0)
}

// entryNotes are the lines of that note, wrapped to innerW cells.
func (m Model) entryNotes(innerW int) []NotesRow {
	if m.selPlaylist >= len(m.state.Playlists) {
		return nil
	}
	pl := m.state.Playlists[m.selPlaylist]
	var notes []NotesRow
	switch {
	case pl.Err != "":
		for _, l := range cells.Wrap("✗ "+cells.Display(pl.Err), innerW) {
			notes = append(notes, NotesRow{Line: cells.CellText(l), LineStyle: CellStyle{Fg: look.Colours.Error}})
		}
	case !pl.Listed:
		notes = []NotesRow{{Line: "listing playlist...", LineStyle: CellStyle{Fg: look.Colours.Dim}}}
	case len(m.visibleEntries()) == 0 && m.filterFailed:
		notes = []NotesRow{{Line: "no failed entries", LineStyle: CellStyle{Fg: look.Colours.Dim}}}
	}
	return notes
}

// footer is what the output pane keeps for the selected entry under its lines.
type outputFooter struct {
	lines []string // empty, or a blank line and then one
	bar   *barInfo
}

type barInfo struct {
	info    string
	percent float64
}

func (m Model) outputFooter() outputFooter {
	var f outputFooter
	en := m.selectedEntry()
	if en == nil || m.focus == paneLists {
		return f
	}
	switch en.State {
	case engine.StateDownloading, engine.StateMerging:
		info := runstate.ProgressLine(en)
		if en.State == engine.StateMerging {
			info = "merging streams with ffmpeg"
		}
		percent := en.Percent
		if en.State == engine.StateMerging || en.Total == 0 {
			percent = 100
		}
		f.bar = &barInfo{info: info, percent: percent}
		f.lines = []string{"", look.Yellow.Render(info)}
	case engine.StateFailed:
		f.lines = []string{"", look.Red.Render("✗ " + cells.Display(en.Err))}
	case engine.StateDone:
		what := "original " + en.TargetID
		if en.Official {
			what = "official video " + en.TargetID
		} else if en.Better {
			what = "better quality video " + en.TargetID
		}
		f.lines = []string{"", look.Green.Render("✓ downloaded " + what)}
	}
	return f
}

// outputWindow are the rows of the output pane, w cells wide and h high: the last h rows
// above the scroll, with a long log line wrapped onto several rows instead of cut. The scroll
// still counts log lines, so a line is never half scrolled away; the second value is the
// index of the log line the first row belongs to.
func (m Model) outputWindow(w, h int) ([]string, int) {
	lines := m.outputLines()
	end := max(0, len(lines)-m.outScroll)
	h = max(h, 1)

	// Built from the bottom up, so only the lines that show are wrapped however long the log is.
	rows := make([]string, 0, h)
	start := end
	for i := end - 1; i >= 0 && len(rows) < h; i-- {
		start = i
		wrapped := cells.WrapLine(cells.Display(lines[i]), w)
		if len(wrapped) == 0 {
			wrapped = []string{""}
		}
		for j := len(wrapped) - 1; j >= 0 && len(rows) < h; j-- {
			rows = append(rows, wrapped[j])
		}
	}
	slices.Reverse(rows)
	return rows, start
}

// leftWidth is how wide the column of lists is in a window w cells wide: all of it when
// the window is narrow, most of it for the output when it is wide, and none when the
// output has the whole screen.
func (m Model) leftWidth(w int) int {
	switch {
	case m.fullLog:
		return 0
	case w < minWideWidth:
		return w
	}
	if m.leftW > 0 {
		return max(minPane, min(m.leftW, w-minPane*2))
	}
	return max(30, w*26/100)
}

// columns is whether the playlists, the entries and the output stand side by side.
func (m Model) columns(w int) bool { return !m.fullLog && w >= minWideWidth }

// midWidth is how wide the column of entries is, in the columns layout.
func (m Model) midWidth(w int) int {
	left := m.leftWidth(w)
	if m.midW > 0 {
		return max(minPane, min(m.midW, w-left-minPane))
	}
	return max(32, w*30/100)
}
