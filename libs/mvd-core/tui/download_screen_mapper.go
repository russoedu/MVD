package tui

import (
	"fmt"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/viewport"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/runstate"
)

// downloadProps fills the properties of the Download screen, drawn from its design, for
// a window of w by h cells. The design owns where everything goes; what is here is what
// the screen shows. The rows are windowed to the room the design gives them.
func (m model) downloadProps(w, h int) downloadpage.Props {
	p := downloadpage.Default()
	p.Palette = downloadpage.Palette(look.CurrentPalette())
	// A design shows its optional parts by default; only what applies is shown.
	p.ShowProgress, p.ShowNote, p.ConfirmQuit, p.ShowStatus, p.ShowKeys = false, false, false, false, false

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

	// The design previews with a footer; the app has one only when there is something for it.
	p.FooterLines = 0
	footer := m.outputFooter()
	if !showOutput {
		footer = outputFooter{}
	}
	if len(footer.lines) > 0 {
		p.FooterLines = 2
	}

	// Place the panes first, so the rows can be cut to the room there is.
	p.Playlists = []downloadpage.PlaylistsRow{{}}
	p.Entries = []downloadpage.EntriesRow{{}}
	p.Output = []downloadpage.OutputRow{{}}
	frame := downloadpage.Layout(p, w, h)
	region := func(name string) downloadpage.Region { r, _ := frame.RegionNamed(name); return r }

	// Playlists.
	pl := region("Playlists rows")
	p.Playlists = viewport.Rows(m.playlistRows(), m.selPlaylist, pl.H)
	p.PlaylistsRight = fmt.Sprintf("%d", len(m.state.Playlists))
	p.PlaylistsBorder = m.borderColour(m.focus == paneLists)

	// Entries.
	en := region("Entries rows")
	title, right, entries, notes, sel := m.entryRows(en.W)
	p.Entries = viewport.Rows(entries, sel, en.H)
	p.Notes, p.HasNotes = notes, 0
	if len(notes) > 0 {
		p.HasNotes = 1
	}
	p.EntriesTitle, p.EntriesRight = title, right
	p.EntriesBorder = m.borderColour(m.focus == paneEntries)

	// Output.
	out := region("Output rows")
	p.Output = nil
	for _, line := range m.outputWindow(out.H) {
		p.Output = append(p.Output, downloadpage.OutputRow{Line: cells.CellText(line)})
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
		inner := region("Output box").W - 2
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
	p.Hints = hintsOf[downloadpage.HintsHint](m.hints())
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

// entryRows are the rows of the entries pane for the selected playlist, with the title
// and the label of the pane and the row that is selected.
func (m model) entryRows(innerW int) (title, right string, rows []downloadpage.EntriesRow, notes []downloadpage.NotesRow, sel int) {
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

	title = "Entries"
	if m.filterFailed {
		title = "Entries (failed only)"
	}
	if m.selPlaylist < len(m.state.Playlists) {
		pl := m.state.Playlists[m.selPlaylist]
		right = fmt.Sprintf("%d", len(ids))
		switch {
		case pl.Err != "":
			for _, l := range cells.Wrap("✗ "+cells.Display(pl.Err), innerW) {
				notes = append(notes, downloadpage.NotesRow{Line: cells.CellText(l), LineStyle: downloadpage.CellStyle{Fg: look.Colours.Error}})
			}
		case !pl.Listed:
			notes = []downloadpage.NotesRow{{Line: "listing playlist...", LineStyle: downloadpage.CellStyle{Fg: look.Colours.Dim}}}
		case len(ids) == 0 && m.filterFailed:
			notes = []downloadpage.NotesRow{{Line: "no failed entries", LineStyle: downloadpage.CellStyle{Fg: look.Colours.Dim}}}
		}
	}
	return title, right, rows, notes, sel
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

// outputWindow are the lines of the output pane, the last h of them above the scroll.
func (m model) outputWindow(h int) []string {
	lines := m.outputLines()
	end := max(0, len(lines)-m.outScroll)
	start := max(0, end-max(h, 1))
	out := make([]string, 0, h)
	for _, l := range lines[start:end] {
		out = append(out, cells.Display(l))
	}
	return out
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
