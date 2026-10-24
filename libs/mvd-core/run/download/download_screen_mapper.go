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
		p.Columns, p.LeftReserve, p.DetailsHeight = 1, m.reserve, m.detailsHeight(h)
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
	if !showOutput || !m.technical {
		footer = outputFooter{}
	}
	if len(footer.lines) > 0 {
		p.FooterLines = 2
	}

	// Playlists.
	p.PlaylistsRows = func(w, h int) ([]PlaylistsRow, int) {
		return scrolled(m.playlistRows(w), m.playlistsTop, h)
	}
	p.PlaylistsTitle = counted(len(m.state.Playlists), "link")
	p.PlaylistsRight = ""
	p.PlaylistsFocus = m.focus == paneLists

	// Entries.
	p.EntriesTitle, p.EntriesRight = m.entryHeader()
	p.EntriesRows = func(w, h int) ([]EntriesRow, int) {
		rows, _ := m.entryRows(w)
		return scrolled(rows, m.entriesTop, h)
	}
	// The songs are a table, with a row of headings, whenever there are songs to show.
	p.EntriesTable = m.songsTable
	if len(m.visibleEntries()) > 0 && !m.entryNotesShown() {
		p.TableHeader = 1
	}
	p.HasNotes = 0
	if m.entryNotesShown() {
		p.HasNotes = 1
	}
	p.NotesRows = func(w, _ int) ([]NotesRow, int) { return m.entryNotes(w), 0 }
	p.EntriesFocus = m.focus == paneEntries

	// Output.
	p.OutputRows = func(w, h int) ([]OutputRow, int) {
		if !m.technical {
			return m.detailRows(w), 0
		}
		lines, first := m.outputWindow(w, h)
		if len(m.outputLines()) == 0 {
			return emptyOutputRows(m.emptyOutputText(), w), 0
		}
		rows := make([]OutputRow, 0, len(lines))
		for _, line := range lines {
			rows = append(rows, OutputRow{Line: cells.CellText(line)})
		}
		return rows, first
	}
	p.OutputTitle = m.outputTitle()
	p.OutputRight = ""
	if m.technical {
		p.OutputRight = "technical log"
		if m.outScroll > 0 {
			p.OutputRight = fmt.Sprintf("↑%d", m.outScroll)
		}
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
	case m.swapping:
		p.ShowStatus = true
		p.Status = " " + m.swap.View()
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
	elapsed := runstate.HumanDuration(m.state.Elapsed())

	right = elapsed
	if m.editable {
		right = "" // nothing is running: there is no time to count
	} else if m.paused {
		right = "PAUSED · " + elapsed
	} else if m.state.Idle && t.Planned > 0 {
		right = "CHOOSE WHAT TO DOWNLOAD · " + elapsed
	} else if m.state.Idle {
		right = "FINISHED · " + elapsed
	} else if t.Downloading > 0 {
		right = downloadingSummary(t) + " · " + elapsed
	}

	// What matters is always there, in everyday words; the rest only when it has something to say.
	parts := []string{
		look.Text.Render(fmt.Sprintf("Waiting %d", t.Queued)),
		look.Yellow.Render(fmt.Sprintf("%s Downloading %d", look.SpinnerFrames[m.spin], t.Running)),
		look.Green.Render(fmt.Sprintf("✓ Done %d", t.Done)),
		look.Red.Render(fmt.Sprintf("✗ Failed %d", t.Failed+m.unreadableLinks())),
	}
	if t.Running == 0 || m.state.Idle || m.paused {
		// Nothing moves, so nothing spins.
		parts[1] = look.Dim.Render(fmt.Sprintf("  Downloading %d", t.Running))
	}
	if t.Planned > 0 {
		parts = append(parts, look.Yellow.Render(fmt.Sprintf("? To choose %d", t.Planned)))
	}
	for _, extra := range []struct {
		n     int
		text  string
		style func(...string) string
	}{
		{t.Official, "⇄ Official video", look.Cyan.Render},
		{t.Better, "↑ Better quality", look.Cyan.Render},
		{t.Duplicate, "≡ Already had", look.Dim.Render},
		{t.NotFound, "∅ No official video", look.Dim.Render},
		{t.Skipped, "– Rejected", look.Dim.Render},
		{t.Retried, "↻ Tried again", look.Cyan.Render},
	} {
		if extra.n > 0 {
			parts = append(parts, extra.style(fmt.Sprintf("%s %d", extra.text, extra.n)))
		}
	}
	return strings.Join(parts, "   "), right
}

// playlistRows are the rows of the playlists pane: the design places the cursor, the
// title, the count and the glyph; the app says what each says and the colour of the glyph.
func (m Model) playlistRows(w int) []PlaylistsRow {
	if len(m.state.Playlists) == 0 {
		return emptyPlaylistRows(w)
	}
	rows := make([]PlaylistsRow, 0, len(m.state.Playlists))
	for i, pl := range m.state.Playlists {
		finished, _, _ := m.state.PlaylistTally(pl)
		finished -= m.waitingIn(pl) // a song that waits to be chosen is not done
		count := fmt.Sprintf("%d/%d", finished, len(pl.Entries))
		if pl.Err != "" {
			count = "can't read"
		} else if !pl.Listed {
			count = "reading"
		}
		glyph, colour := m.playlistGlyph(pl)
		title := cells.CellText(cells.Display(m.playlistName(pl)))
		row := PlaylistsRow{
			Mark: " ", Title: title, Count: count, Glyph: glyph,
			GlyphStyle: CellStyle{Fg: colour},
		}
		if m.editable {
			row.Mark, row.Count, row.Glyph = "≡", "✎", "×"
			row.GlyphStyle, row.CountStyle = CellStyle{Fg: look.Colours.Error}, CellStyle{Fg: look.Colours.Focus}
		}
		if !m.editable && m.linkEditing && i == m.selPlaylist {
			row.Count, row.Glyph = "✎", "×"
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
		for _, line := range wrapText(m.emptyEntriesText(), w-6) {
			rows = append(rows, EntriesRow{Glyph: " ", Sp: " ", Title: line, TitleStyle: CellStyle{Fg: look.Colours.Dim}})
		}
		return rows, 0
	}
	sel = viewport.Clamp(m.selEntry, 0, max(0, len(ids)-1))

	cols := m.tableColumns(w)
	for i, id := range ids {
		texts, styles := m.songCells(m.state.Entries[id], cols, i == sel)
		row := EntriesRow{Cells: texts, CellStyles: styles}
		if i == sel {
			// Whether the pane has the focus is the design's: an unfocused pane draws it as text.
			row.Style = RowSelected
		}
		rows = append(rows, row)
	}
	return rows, sel
}

// entryHeader are the title and the label of the entries pane.
func (m Model) entryHeader() (title, right string) {
	n := 0
	if m.selPlaylist < len(m.state.Playlists) {
		n = len(m.visibleEntries())
	}
	title = counted(n, "song")
	if m.filterActive() && m.selPlaylist < len(m.state.Playlists) {
		title = fmt.Sprintf("%d of %s", n, counted(len(m.state.Playlists[m.selPlaylist].Entries), "song"))
	}

	return title, ""
}

// entryNotesShown is whether the entries pane has a note to show instead of rows: the
// playlist has an error, is still being listed, or has no failed entries to show.
func (m Model) entryNotesShown() bool {
	if m.selPlaylist >= len(m.state.Playlists) {
		return false
	}
	pl := m.state.Playlists[m.selPlaylist]
	return pl.Err != "" || !pl.Listed || (m.filterActive() && len(m.visibleEntries()) == 0)
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
		for _, l := range cells.Wrap("✗ "+cells.Display(runstate.PlainError(pl.Err)), innerW) {
			notes = append(notes, NotesRow{Line: cells.CellText(l), LineStyle: CellStyle{Fg: look.Colours.Error}})
		}
	case !pl.Listed:
		notes = []NotesRow{{Line: "reading this link...", LineStyle: CellStyle{Fg: look.Colours.Dim}}}
	case len(m.visibleEntries()) == 0 && m.filterActive():
		notes = []NotesRow{{Line: "no songs with these statuses: change the filter above", LineStyle: CellStyle{Fg: look.Colours.Dim}}}
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
		if en.State == engine.StateMerging {
			percent = 100 // the parts are all in: the files are being joined
		} else if en.Total == 0 {
			percent = 0 // the size is not known yet: an empty bar, not a full one
		}
		f.bar = &barInfo{info: info, percent: percent}
		f.lines = []string{"", look.Yellow.Render(info)}
	case engine.StateFailed:
		if m.technical {
			f.lines = []string{"", look.Red.Render("✗ " + cells.Display(runstate.PlainError(en.Err)))}
		}
	case engine.StateDone:
		if !m.technical {
			break // the details say it
		}
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

// detailsHeight is how tall the details pane under the songs is in a window h cells tall: enough
// for the account of a song, or half of what there is for the log of the tool.
func (m Model) detailsHeight(h int) int {
	room := max(h-4, 6)
	if m.technical {
		return max(6, room/2)
	}

	base := min(7, max(4, room/3))
	if m.anyWaiting() {
		// The buttons of a song that waits to be chosen sit on the last row.
		return min(max(base, choiceButtonsHeight), max(room-5, base))
	}

	return base
}

// emptyPlaylistRows is what the playlists pane says while there is nothing in it: what to do
// first, how, and what works.
func emptyPlaylistRows(w int) []PlaylistsRow {
	var rows []PlaylistsRow
	say := func(text string, style CellStyle) {
		for _, line := range wrapText(text, w-6) {
			rows = append(rows, PlaylistsRow{Mark: " ", Sp: " ", Title: line, TitleStyle: style})
		}
	}
	blank := func() { rows = append(rows, PlaylistsRow{Mark: " ", Sp: " ", Title: " "}) }
	say("Add the songs you want as music videos.", CellStyle{Fg: look.Colours.Text, Bold: true})
	blank()
	say("Click + Add link below, press Ctrl+L, or paste links with Ctrl+V.", CellStyle{Fg: look.Colours.Text})
	blank()
	say("Works with YouTube videos and playlists, and with Spotify and Apple Music playlists.", CellStyle{Fg: look.Colours.Dim})
	return rows
}

// emptyEntriesText is what the entries pane says while it has no songs to show: that there is
// nothing yet, or that a list has been made and not looked up, and how to do it.
func (m Model) emptyEntriesText() string {
	switch {
	case len(m.state.Playlists) == 0:
		return "The songs of each link show here, once you add one."
	case m.editable:
		return "Nothing loaded yet. Press Load playlists (Ctrl+R) to see the songs of each playlist, then choose what to download."
	case m.selPlaylist < len(m.state.Playlists) && !m.state.Playlists[m.selPlaylist].Listed:
		return "Reading the songs of this link…"
	}
	return "No songs to show here."
}

// emptyOutputText is what the output pane says when nothing was written for what is selected:
// before the run it says what will show; for a playlist that could not be read, why there is
// none; otherwise that this one wrote nothing.
func (m Model) emptyOutputText() string {
	switch {
	case m.selPlaylist < len(m.state.Playlists) && m.state.Playlists[m.selPlaylist].Err != "":
		return "Nothing to show: this link could not be read."
	case m.editable || len(m.state.Playlists) == 0:
		return "Nothing yet. What happens to each song shows here while it downloads."
	}
	return "Nothing to show for this one."
}

// emptyOutputRows is the text of the output pane, wrapped and dim, before anything has been written.
func emptyOutputRows(text string, w int) []OutputRow {
	var rows []OutputRow
	for _, line := range wrapText(text, w-2) {
		rows = append(rows, OutputRow{Line: cells.CellText(line), LineStyle: CellStyle{Fg: look.Colours.Dim}})
	}
	return rows
}

// unreadableLinks is how many links could not be read at all: no song of them is counted, and for
// the person it is a failure all the same.
func (m Model) unreadableLinks() int {
	n := 0
	for _, pl := range m.state.Playlists {
		if pl.Err != "" {
			n++
		}
	}

	return n
}
