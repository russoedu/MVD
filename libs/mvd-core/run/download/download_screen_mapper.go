package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsnapshot"
	"fmt"
	"slices"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/keybar"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
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
	if m.session.FullLog {
		p.FullLog = 1
	}
	showOutput := m.session.FullLog || w >= minWideWidth
	p.PlaylistCount = len(m.session.State.Playlists)

	p.Counters, p.HeaderRight = m.headerTexts()
	p.Counters = cells.Truncate(p.Counters, w-2)

	// The app has a footer under the output only when there is something for it.
	footer := m.outputFooter()
	if !showOutput || !m.session.Technical {
		footer = outputFooter{}
	}
	if len(footer.lines) > 0 {
		p.FooterLines = 2
	}

	// Playlists.
	p.PlaylistsRows = func(w, h int) ([]PlaylistsRow, int) {
		return scrolled(m.playlistRows(w), m.playlistsTop, h)
	}
	p.PlaylistsTitle = counted(len(m.session.State.Playlists), "link")
	p.PlaylistsRight = ""
	p.PlaylistsFocus = m.session.Pane == paneLists

	// Entries.
	p.EntriesTitle, p.EntriesRight = m.entryHeader()
	p.EntriesRows = func(w, h int) ([]EntriesRow, int) {
		rows, _ := m.entryRows(w)
		return scrolled(rows, m.entriesTop, h)
	}
	// The songs are a table, with a row of headings, whenever there are songs to show.
	p.EntriesTable = m.songsTable
	if len(m.session.VisibleEntries()) > 0 && !m.entryNotesShown() {
		p.TableHeader = 1
	}
	p.HasNotes = 0
	if m.entryNotesShown() {
		p.HasNotes = 1
	}
	p.NotesRows = func(w, _ int) ([]NotesRow, int) { return m.entryNotes(w), 0 }
	p.EntriesFocus = m.session.Pane == paneEntries

	// Output.
	p.OutputRows = func(w, h int) ([]OutputRow, int) {
		if !m.session.Technical {
			return m.detailRows(w), 0
		}
		lines, first := m.outputWindow(w, h)
		if len(m.session.OutputLines()) == 0 {
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
	if m.session.Technical {
		p.OutputRight = "technical log"
		if m.session.OutScroll > 0 {
			p.OutputRight = fmt.Sprintf("↑%d", m.session.OutScroll)
		}
	}
	p.OutputFocus = m.session.Pane == paneOutput

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
	case m.session.ConfirmQuit:
		p.ConfirmQuit = true
		p.Confirm = look.Red.Render(" Downloads are still running. Quit? ") + look.Key.Render("y") + look.Dim.Render("/") + look.Key.Render("n")
	case m.session.Swapping:
		p.ShowStatus = true
		p.Status = " " + m.swap.View()
	case m.session.Notice != "" && time.Now().Before(m.session.NoticeUntil):
		p.ShowStatus = true
		p.Status = " " + m.session.Notice
	default:
		// The keys are for a run; while the list is only being made there is nothing to press.
		p.ShowKeys = !m.session.Editable
	}
	return p
}

// headerTexts are the counters inside the header box and the label on its top edge.
func (m Model) headerTexts() (counters, right string) {
	h := downloadsnapshot.HeaderOf(m.session)
	parts := make([]string, 0, len(h.Counters))
	for _, c := range h.Counters {
		parts = append(parts, m.counterText(c))
	}

	return strings.Join(parts, "   "), h.Right
}

// counterGlyphs are the marks in front of the counters of the header.
var counterGlyphs = map[downloadsnapshot.CounterKind]string{
	downloadsnapshot.CounterDone: "✓", downloadsnapshot.CounterFailed: "✗", downloadsnapshot.CounterOfficialFound: "✓",
	downloadsnapshot.CounterOriginalOnly: "?", downloadsnapshot.CounterAccepted: "▸", downloadsnapshot.CounterOfficialVideo: "⇄",
	downloadsnapshot.CounterBetterQuality: "↑", downloadsnapshot.CounterAlreadyHad: "≡", downloadsnapshot.CounterNoOfficial: "∅",
	downloadsnapshot.CounterRejected: "–", downloadsnapshot.CounterTriedAgain: "↻",
}

// counterText is one counter of the header, in the colour of its tone.
func (m Model) counterText(c downloadsnapshot.Counter) string {
	style := toneStyle(c.Tone)
	switch {
	case c.Kind == downloadsnapshot.CounterWaiting:
		return style(fmt.Sprintf("Waiting %d", c.Count))
	case c.Kind == downloadsnapshot.CounterDownloading && c.Moving:
		return style(fmt.Sprintf("%s Downloading %d", look.SpinnerFrames[m.spin], c.Count))
	case c.Kind == downloadsnapshot.CounterDownloading:
		return style(fmt.Sprintf("  Downloading %d", c.Count))
	}

	return style(fmt.Sprintf("%s %s %d", counterGlyphs[c.Kind], c.Label, c.Count))
}

// playlistRows are the rows of the playlists pane: the design places the cursor, the
// title, the count and the glyph; the app says what each says and the colour of the glyph.
func (m Model) playlistRows(w int) []PlaylistsRow {
	if len(m.session.State.Playlists) == 0 {
		return emptyPlaylistRows(w)
	}
	snapshot := downloadsnapshot.PlaylistRowsOf(m.session)
	rows := make([]PlaylistsRow, 0, len(snapshot))
	for i, pl := range snapshot {
		glyph, colour := m.glyphOf(pl.Glyph)
		title := cells.CellText(cells.Display(pl.Name))
		row := PlaylistsRow{
			Mark: " ", Title: title, Count: pl.Count, Glyph: glyph,
			GlyphStyle: CellStyle{Fg: colour},
		}
		if m.session.Editable {
			row.Mark, row.Count, row.Glyph = "≡", "✎", "×"
			row.GlyphStyle, row.CountStyle = CellStyle{Fg: look.Colours.Error}, CellStyle{Fg: look.Colours.Focus}
		}
		if !m.session.Editable && m.session.LinkEditing && i == m.session.SelPlaylist {
			row.Count, row.Glyph = "✎", "×"
			row.GlyphStyle, row.CountStyle = CellStyle{Fg: look.Colours.Error}, CellStyle{Fg: look.Colours.Focus}
		}
		if m.dragRowOn && i == m.dropRow {
			row.Mark = "▶"
		}
		if i == m.session.SelPlaylist {
			// 6 cells are the mark, the gaps, the glyph and the end of the row.
			row.Style = RowSelected
			if !m.session.Editable {
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
	ids := m.session.VisibleEntries()
	if len(ids) == 0 {
		var rows []EntriesRow
		for _, line := range wrapText(m.emptyEntriesText(), w-6) {
			rows = append(rows, EntriesRow{Glyph: " ", Sp: " ", Title: line, TitleStyle: CellStyle{Fg: look.Colours.Dim}})
		}
		return rows, 0
	}
	sel = viewport.Clamp(m.session.SelEntry, 0, max(0, len(ids)-1))

	cols := m.tableColumns(w)
	for i, id := range ids {
		texts, styles := m.songCells(m.session.State.Entries[id], cols, i == sel)
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
	if m.session.SelPlaylist < len(m.session.State.Playlists) {
		n = len(m.session.VisibleEntries())
	}
	title = counted(n, "song")
	if m.session.Filter.Active() && m.session.SelPlaylist < len(m.session.State.Playlists) {
		title = fmt.Sprintf("%d of %s", n, counted(len(m.session.State.Playlists[m.session.SelPlaylist].Entries), "song"))
	}

	return title, ""
}

// entryNotesShown is whether the entries pane has a note to show instead of rows: the
// playlist has an error, is still being listed, or has no failed entries to show.
func (m Model) entryNotesShown() bool {
	if m.session.SelPlaylist >= len(m.session.State.Playlists) {
		return false
	}
	pl := m.session.State.Playlists[m.session.SelPlaylist]
	return pl.Err != "" || !pl.Listed || (m.session.Filter.Active() && len(m.session.VisibleEntries()) == 0)
}

// entryNotes are the lines of that note, wrapped to innerW cells.
func (m Model) entryNotes(innerW int) []NotesRow {
	if m.session.SelPlaylist >= len(m.session.State.Playlists) {
		return nil
	}
	pl := m.session.State.Playlists[m.session.SelPlaylist]
	var notes []NotesRow
	switch {
	case pl.Err != "":
		for _, l := range cells.Wrap("✗ "+cells.Display(runstate.PlainError(pl.Err)), innerW) {
			notes = append(notes, NotesRow{Line: cells.CellText(l), LineStyle: CellStyle{Fg: look.Colours.Error}})
		}
	case !pl.Listed:
		notes = []NotesRow{{Line: "reading this link...", LineStyle: CellStyle{Fg: look.Colours.Dim}}}
	case len(m.session.VisibleEntries()) == 0 && m.session.Filter.Active():
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
	foot := downloadsnapshot.FooterOf(m.session)
	if foot.None {
		return f
	}
	if foot.HasBar {
		f.bar = &barInfo{info: foot.Info, percent: foot.Percent}
	}
	f.lines = []string{"", toneStyle(foot.Tone)(cells.Display(foot.Info))}

	return f
}

// outputWindow are the rows of the output pane, w cells wide and h high: the last h rows
// above the scroll, with a long log line wrapped onto several rows instead of cut. The scroll
// still counts log lines, so a line is never half scrolled away; the second value is the
// index of the log line the first row belongs to.
func (m Model) outputWindow(w, h int) ([]string, int) {
	lines := m.session.OutputLines()
	end := max(0, len(lines)-m.session.OutScroll)
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
	case m.session.FullLog:
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
func (m Model) columns(w int) bool { return !m.session.FullLog && w >= minWideWidth }

// detailsHeight is how tall the details pane under the songs is in a window h cells tall: enough
// for the account of a song, or half of what there is for the log of the tool.
func (m Model) detailsHeight(h int) int {
	room := max(h-4, 6)
	if m.session.Technical {
		return max(6, room/2)
	}

	base := min(7, max(4, room/3))
	if m.session.AnyWaiting() {
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

// emptyEntriesText is what the entries pane says while it has no songs to show.
func (m Model) emptyEntriesText() string { return downloadsnapshot.EmptyEntriesText(m.session) }

// emptyOutputText is what the output pane says when nothing was written for what is selected.
func (m Model) emptyOutputText() string { return downloadsnapshot.EmptyOutputText(m.session) }

// emptyOutputRows is the text of the output pane, wrapped and dim, before anything has been written.
func emptyOutputRows(text string, w int) []OutputRow {
	var rows []OutputRow
	for _, line := range wrapText(text, w-2) {
		rows = append(rows, OutputRow{Line: cells.CellText(line), LineStyle: CellStyle{Fg: look.Colours.Dim}})
	}
	return rows
}

