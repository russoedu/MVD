package tui

import (
	"fmt"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/keybar"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/internal/cuppa"
	"youtube-downloader/libs/mvd-core/look"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/viewport"
)

// downloadProps fills the properties of the Download screen, drawn from its design, for
// a window of w by h cells. The design owns where everything goes; what is here is what
// the screen shows. The rows come as functions of the room the design gives them, so the
// screen cuts them to fit and a click is matched with the page's own At functions.
func (m model) downloadProps(w, h int) downloadpage.Props {
	p := downloadpage.Default()
	p.Palette = downloadpage.Palette(look.CurrentPalette())

	// The sizes the design cannot know: how wide the left column is.
	p.LeftWidth = m.leftWidth(w)
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
	p.PlaylistsRows = func(_, h int) ([]downloadpage.PlaylistsRow, int) {
		return cuppa.Window(m.playlistRows(), m.selPlaylist, h)
	}
	p.PlaylistsRight = fmt.Sprintf("%d", len(m.state.Playlists))
	p.PlaylistsBorder = m.borderColour(m.focus == paneLists)

	// Entries.
	p.EntriesTitle, p.EntriesRight = m.entryHeader()
	p.EntriesRows = func(_, h int) ([]downloadpage.EntriesRow, int) {
		rows, sel := m.entryRows()
		return cuppa.Window(rows, sel, h)
	}
	p.HasNotes = 0
	if m.entryNotesShown() {
		p.HasNotes = 1
	}
	p.NotesRows = func(w, _ int) ([]downloadpage.NotesRow, int) { return m.entryNotes(w), 0 }
	p.EntriesBorder = m.borderColour(m.focus == paneEntries)

	// Output.
	p.OutputRows = func(_, h int) ([]downloadpage.OutputRow, int) {
		lines, first := m.outputWindow(h)
		rows := make([]downloadpage.OutputRow, 0, len(lines))
		for _, line := range lines {
			rows = append(rows, downloadpage.OutputRow{Line: cells.CellText(line)})
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
	p.OutputBorder = m.borderColour(m.focus == paneOutput)

	// The footer of the output pane: a progress bar, or a note.
	if len(footer.lines) > 0 {
		outputBox, _ := downloadpage.Layout(p, w, h).RegionNamed("Output box")
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
	p.Hints = keybar.Of[downloadpage.HintsHint](m.hints())
	switch {
	case m.confirmQuit:
		p.ConfirmQuit = true
		p.Confirm = look.Red.Render(" Downloads are still running. Quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	case m.status != "" && time.Now().Before(m.statusUntil):
		p.ShowStatus = true
		p.Status = " " + m.status
	default:
		p.ShowKeys = true
	}
	return p
}

func (m model) borderColour(focused bool) string {
	if focused {
		return look.Colours.Focus
	}
	return look.Colours.Dim
}

// headerTexts are the counters inside the header box and the label on its top edge.
func (m model) headerTexts() (counters, right string) {
	t := m.state.Tally()
	elapsed := runstate.HumanDuration(time.Since(m.state.Started))

	right = elapsed
	if m.state.Idle {
		right = "FINISHED · " + elapsed
	} else if t.Downloading > 0 {
		right = downloadingSummary(t) + " · " + elapsed
	}

	parts := []string{
		look.Text.Render(fmt.Sprintf("Queue %d", t.Queued)),
		look.Yellow.Render(fmt.Sprintf("%s Running %d", spinnerFrames[m.spin], t.Running)),
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
func (m model) playlistRows() []downloadpage.PlaylistsRow {
	rows := make([]downloadpage.PlaylistsRow, 0, len(m.state.Playlists))
	for i, pl := range m.state.Playlists {
		finished, _, _ := m.state.PlaylistTally(pl)
		count := fmt.Sprintf("%d/%d", finished, len(pl.Entries))
		if pl.Err != "" {
			count = "error"
		} else if !pl.Listed {
			count = "listing"
		}
		glyph, colour := m.playlistGlyph(pl)
		row := downloadpage.PlaylistsRow{
			Mark: " ", Title: cells.CellText(cells.Display(pl.Title)), Count: count, Glyph: glyph,
			GlyphStyle: downloadpage.CellStyle{Fg: colour},
		}
		if i == m.selPlaylist {
			row.Mark, row.Style = "▸", downloadpage.RowSelected
		}
		rows = append(rows, row)
	}
	return rows
}

// entryRows are the rows of the entries pane for the selected playlist and the row that
// is selected.
func (m model) entryRows() (rows []downloadpage.EntriesRow, sel int) {
	ids := m.visibleEntries()
	sel = viewport.Clamp(m.selEntry, 0, max(0, len(ids)-1))

	for i, id := range ids {
		en := m.state.Entries[id]
		glyph, glyphColour := m.entryGlyph(en)
		tag, tagColour := m.entryTag(en)
		row := downloadpage.EntriesRow{
			Glyph: glyph, Idx: fmt.Sprintf("%02d", en.Index), Title: cells.CellText(entryTitle(en)), Tag: tag,
			GlyphStyle: downloadpage.CellStyle{Fg: glyphColour}, TagStyle: downloadpage.CellStyle{Fg: tagColour},
		}
		switch {
		case i == sel && m.focus == paneEntries:
			row.Style = downloadpage.RowSelected
		case i == sel:
			// A selected row in a pane that does not have the focus is only coloured.
			accent := downloadpage.CellStyle{Fg: look.Colours.Accent}
			row.SpStyle, row.IdxStyle, row.Sp2Style, row.TitleStyle, row.Sp3Style = accent, accent, accent, accent, accent
		}
		rows = append(rows, row)
	}
	return rows, sel
}

// entryHeader are the title and the label of the entries pane.
func (m model) entryHeader() (title, right string) {
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
func (m model) entryNotesShown() bool {
	if m.selPlaylist >= len(m.state.Playlists) {
		return false
	}
	pl := m.state.Playlists[m.selPlaylist]
	return pl.Err != "" || !pl.Listed || (m.filterFailed && len(m.visibleEntries()) == 0)
}

// entryNotes are the lines of that note, wrapped to innerW cells.
func (m model) entryNotes(innerW int) []downloadpage.NotesRow {
	if m.selPlaylist >= len(m.state.Playlists) {
		return nil
	}
	pl := m.state.Playlists[m.selPlaylist]
	var notes []downloadpage.NotesRow
	switch {
	case pl.Err != "":
		for _, l := range cells.Wrap("✗ "+cells.Display(pl.Err), innerW) {
			notes = append(notes, downloadpage.NotesRow{Line: cells.CellText(l), LineStyle: downloadpage.CellStyle{Fg: look.Colours.Error}})
		}
	case !pl.Listed:
		notes = []downloadpage.NotesRow{{Line: "listing playlist...", LineStyle: downloadpage.CellStyle{Fg: look.Colours.Dim}}}
	case len(m.visibleEntries()) == 0 && m.filterFailed:
		notes = []downloadpage.NotesRow{{Line: "no failed entries", LineStyle: downloadpage.CellStyle{Fg: look.Colours.Dim}}}
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

func (m model) outputFooter() outputFooter {
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

// outputWindow are the lines of the output pane, the last h of them above the scroll, and
// the index of the first of them among all the lines.
func (m model) outputWindow(h int) ([]string, int) {
	lines := m.outputLines()
	end := max(0, len(lines)-m.outScroll)
	start := max(0, end-max(h, 1))
	out := make([]string, 0, h)
	for _, l := range lines[start:end] {
		out = append(out, cells.Display(l))
	}
	return out, start
}

// leftWidth is how wide the column of lists is in a window w cells wide: all of it when
// the window is narrow, most of it for the output when it is wide, and none when the
// output has the whole screen.
func (m model) leftWidth(w int) int {
	switch {
	case m.fullLog:
		return 0
	case w < minWideWidth:
		return w
	}
	return max(36, w*38/100)
}
