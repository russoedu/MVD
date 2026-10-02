package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Palette taken from the logo.
var (
	colMagenta = lipgloss.Color("#ff007f")
	colCyan    = lipgloss.Color("#00f0ff")
	colYellow  = lipgloss.Color("#ffe600")
	colGreen   = lipgloss.Color("#3ddc84")
	colRed     = lipgloss.Color("#ff4d4d")
	colDim     = lipgloss.Color("#6b7280")
	colText    = lipgloss.Color("#e5e7eb")

	styTitle    = lipgloss.NewStyle().Bold(true).Foreground(colMagenta)
	styBorder   = lipgloss.NewStyle().Foreground(colDim)
	styBorderOn = lipgloss.NewStyle().Foreground(colCyan)
	styDim      = lipgloss.NewStyle().Foreground(colDim)
	styText     = lipgloss.NewStyle().Foreground(colText)
	styCyan     = lipgloss.NewStyle().Foreground(colCyan)
	styYellow   = lipgloss.NewStyle().Foreground(colYellow)
	styGreen    = lipgloss.NewStyle().Foreground(colGreen)
	styRed      = lipgloss.NewStyle().Foreground(colRed)
	styMagenta  = lipgloss.NewStyle().Foreground(colMagenta)
	stySelected = lipgloss.NewStyle().Background(lipgloss.Color("#3b0f2a")).Foreground(colText).Bold(true)
	styKey      = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	styBarFill  = lipgloss.NewStyle().Foreground(colYellow)
	styBarEmpty = lipgloss.NewStyle().Foreground(colDim)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	paneLists = iota
	paneEntries
	paneOutput
)

const (
	minWideWidth = 100
	headerHeight = 3
	keyBarHeight = 1
)

type (
	evMsg    struct{ ev interface{} }
	evClosed struct{}
	tickMsg  time.Time
)

type tuiModel struct {
	eng    *Engine
	events <-chan interface{}
	state  *runState

	width, height int
	focus         int
	selPlaylist   int
	selEntry      int
	outScroll     int // lines scrolled up from the bottom of the output pane
	follow        bool
	filterFailed  bool
	fullLog       bool
	showHelp      bool
	confirmQuit   bool
	spin          int
	status        string // transient message shown in the key bar
	statusUntil   time.Time
}

func newTUIModel(eng *Engine) tuiModel {
	return tuiModel{
		eng:    eng,
		events: eng.Events(),
		state:  newRunState(eng.Sources()),
		follow: true,
	}
}

func waitEvent(ch <-chan interface{}) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return evClosed{}
		}
		return evMsg{ev}
	}
}

func tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(waitEvent(m.events), tick())
}

// --- update -----------------------------------------------------------------

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.spin = (m.spin + 1) % len(spinnerFrames)
		return m, tick()

	case evMsg:
		id := m.state.apply(msg.ev)
		if st, ok := msg.ev.(EvEntryState); ok && m.follow {
			if st.State == StateDownloading || st.State == StateResolving {
				m.jumpTo(id)
			}
		}
		if lg, ok := msg.ev.(EvLog); ok && m.outScroll > 0 {
			// Keep the view anchored while the user is reading older lines.
			if en := m.selectedEntry(); en != nil && lg.Entry == en.ID {
				m.outScroll++
			}
		}
		return m, waitEvent(m.events)

	case evClosed:
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}
	if key == "ctrl+l" {
		return m, tea.ClearScreen
	}

	if m.confirmQuit {
		switch key {
		case "y", "Y", "enter":
			return m, tea.Quit
		default:
			m.confirmQuit = false
		}
		return m, nil
	}

	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch key {
	case "q", "esc":
		if m.state.Idle {
			return m, tea.Quit
		}
		m.confirmQuit = true
	case "?":
		m.showHelp = true
	case "tab":
		m.focus = (m.focus + 1) % 3
		if m.focus == paneOutput && !m.outputVisible() {
			m.focus = paneLists
		}
	case "shift+tab":
		m.focus = (m.focus + 2) % 3
		if m.focus == paneOutput && !m.outputVisible() {
			m.focus = paneEntries
		}
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "pgup":
		m.move(-10)
	case "pgdown":
		m.move(10)
	case "home":
		m.move(-1 << 20)
	case "end":
		m.move(1 << 20)
	case "enter":
		m.follow = !m.follow
		if m.follow {
			m.outScroll = 0
			if id := m.activeEntry(); id >= 0 {
				m.jumpTo(id)
			}
			m.flash("following active download")
		} else {
			m.flash("follow off")
		}
	case "f":
		m.filterFailed = !m.filterFailed
		m.selEntry = 0
		if m.filterFailed {
			m.flash("showing failed entries only")
		} else {
			m.flash("showing all entries")
		}
	case "l":
		m.fullLog = !m.fullLog
		if m.fullLog {
			m.focus = paneOutput
		} else if m.focus == paneOutput && !m.outputVisible() {
			m.focus = paneEntries
		}
	case "r":
		m.retry()
	}
	return m, nil
}

func (m *tuiModel) flash(s string) {
	m.status = s
	m.statusUntil = time.Now().Add(3 * time.Second)
}

func (m *tuiModel) outputVisible() bool {
	return m.fullLog || m.width >= minWideWidth
}

func (m *tuiModel) move(delta int) {
	switch m.focus {
	case paneLists:
		m.follow = false
		m.selPlaylist = clamp(m.selPlaylist+delta, 0, len(m.state.Playlists)-1)
		m.selEntry = 0
		m.outScroll = 0
	case paneEntries:
		m.follow = false
		n := len(m.visibleEntries())
		m.selEntry = clamp(m.selEntry+delta, 0, n-1)
		m.outScroll = 0
	case paneOutput:
		lines := len(m.outputLines())
		m.outScroll = clamp(m.outScroll-delta, 0, max(0, lines-1))
	}
}

func (m *tuiModel) retry() {
	switch m.focus {
	case paneEntries:
		en := m.selectedEntry()
		if en == nil {
			return
		}
		if en.State != StateFailed {
			m.flash("only failed entries can be retried")
			return
		}
		if m.eng.Retry(en.ID) {
			m.flash(fmt.Sprintf("retrying %02d %s", en.Index, entryTitle(en)))
		}
	default:
		if m.selPlaylist < len(m.state.Playlists) {
			n := m.eng.RetryPlaylist(m.selPlaylist)
			m.flash(fmt.Sprintf("re-queued %d failed entries", n))
		}
	}
}

// visibleEntries returns the entry ids shown in the entries pane.
func (m *tuiModel) visibleEntries() []int {
	if m.selPlaylist >= len(m.state.Playlists) {
		return nil
	}
	pl := m.state.Playlists[m.selPlaylist]
	if !m.filterFailed {
		return pl.Entries
	}
	var out []int
	for _, id := range pl.Entries {
		if m.state.Entries[id].State == StateFailed {
			out = append(out, id)
		}
	}
	return out
}

func (m *tuiModel) selectedEntry() *entryView {
	ids := m.visibleEntries()
	if len(ids) == 0 {
		return nil
	}
	i := clamp(m.selEntry, 0, len(ids)-1)
	return m.state.Entries[ids[i]]
}

// activeEntry returns the id of the first entry currently downloading.
func (m *tuiModel) activeEntry() int {
	for _, en := range m.state.Entries {
		if en != nil && (en.State == StateDownloading || en.State == StateMerging || en.State == StateResolving) {
			return en.ID
		}
	}
	return -1
}

func (m *tuiModel) jumpTo(id int) {
	en := m.state.entry(id)
	if en == nil {
		return
	}
	m.selPlaylist = en.Playlist
	for i, eid := range m.visibleEntries() {
		if eid == id {
			m.selEntry = i
			break
		}
	}
	m.outScroll = 0
}

// --- view -------------------------------------------------------------------

func (m tuiModel) View() string {
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

	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderHelp(), lipgloss.WithWhitespaceChars(" "))
	}
	return view
}

func (m tuiModel) renderHeader(w int) string {
	t := m.state.tally()
	elapsed := humanDuration(time.Since(m.state.Started))

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
	if m.state.Idle {
		parts[1] = styDim.Render(fmt.Sprintf("  Running %d", t.Running))
	}
	counters := strings.Join(parts, "   ")

	return renderBox(title, right, w, headerHeight, []string{counters}, false)
}

func (m tuiModel) renderLeft(w, h int) string {
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

func (m tuiModel) playlistGlyph(pl *playlistView) string {
	if pl.Err != "" {
		return styRed.Render("✗")
	}
	if !pl.Listed {
		return styDim.Render(spinnerFrames[m.spin])
	}
	finished, failed, active := m.state.playlistTally(pl)
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

func (m tuiModel) renderPlaylists(w, h int) string {
	innerW := w - 2
	rows := make([]string, 0, len(m.state.Playlists))
	for i, pl := range m.state.Playlists {
		finished, _, _ := m.state.playlistTally(pl)
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

func (m tuiModel) entryGlyph(en *entryView) string {
	switch en.State {
	case StateQueued:
		return styDim.Render("·")
	case StateResolving:
		return styCyan.Render("⟲")
	case StateDownloading:
		return styYellow.Render(spinnerFrames[m.spin])
	case StateMerging:
		return styYellow.Render("⚙")
	case StateDone:
		return styGreen.Render("✓")
	case StateDuplicate:
		return styDim.Render("≡")
	case StateFailed:
		return styRed.Render("✗")
	}
	return " "
}

func (m tuiModel) entryTag(en *entryView) (string, int) {
	switch en.State {
	case StateResolving:
		return styCyan.Render("res."), 4
	case StateDownloading:
		if en.Total > 0 {
			s := fmt.Sprintf("%3.0f%%", en.Percent)
			return styYellow.Render(s), len(s)
		}
		return styYellow.Render("dl"), 2
	case StateMerging:
		return styYellow.Render("merge"), 5
	case StateDuplicate:
		return styDim.Render("dup"), 3
	case StateFailed:
		return styRed.Render("ERR"), 3
	case StateDone:
		if en.Official {
			return styCyan.Render("⇄"), 1
		}
	}
	return "", 0
}

func (m tuiModel) renderEntries(w, h int) string {
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
func (m tuiModel) outputLines() []string {
	if m.focus == paneLists || m.selectedEntry() == nil {
		if m.selPlaylist < len(m.state.Playlists) {
			return m.state.Playlists[m.selPlaylist].Log
		}
		return nil
	}
	return m.selectedEntry().Log
}

func (m tuiModel) outputTitle() string {
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

func (m tuiModel) renderOutput(w, h int) string {
	innerW := w - 2
	innerH := h - 2
	lines := m.outputLines()

	var footer []string
	if en := m.selectedEntry(); en != nil && m.focus != paneLists {
		switch en.State {
		case StateDownloading, StateMerging:
			footer = append(footer, "", m.renderBar(en, innerW))
		case StateFailed:
			footer = append(footer, "", styRed.Render(truncate("✗ "+display(en.Err), innerW)))
		case StateDone:
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

func (m tuiModel) renderBar(en *entryView, w int) string {
	info := progressLine(en)
	if en.State == StateMerging {
		info = "merging streams with ffmpeg"
	}
	barW := w - ansi.StringWidth(info) - 2
	if barW < 10 {
		return styYellow.Render(truncate(info, w))
	}
	filled := int(float64(barW) * en.Percent / 100)
	if en.State == StateMerging || en.Total == 0 {
		filled = barW
	}
	filled = clamp(filled, 0, barW)
	bar := styBarFill.Render(strings.Repeat("█", filled)) + styBarEmpty.Render(strings.Repeat("░", barW-filled))
	return bar + "  " + styYellow.Render(info)
}

func (m tuiModel) renderKeyBar(w int) string {
	if m.confirmQuit {
		return fit(styRed.Render(" Downloads are still running. Quit? ")+styKey.Render("y")+styDim.Render("/")+styKey.Render("n"), w)
	}
	if m.status != "" && time.Now().Before(m.statusUntil) {
		return fit(" "+styMagenta.Render(m.status), w)
	}
	type kb struct{ k, d string }
	keys := []kb{{"↑↓", "move"}, {"tab", "pane"}, {"⏎", "follow"}, {"f", "failed"}, {"r", "retry"}, {"l", "full log"}, {"?", "help"}, {"q", "quit"}}
	if m.state.Idle {
		keys = []kb{{"↑↓", "move"}, {"tab", "pane"}, {"f", "failed"}, {"r", "retry failed"}, {"l", "full log"}, {"q", "quit"}}
	}
	var b strings.Builder
	b.WriteString(" ")
	for i, k := range keys {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(styKey.Render(k.k) + " " + styDim.Render(k.d))
	}
	if m.state.Idle {
		b.WriteString("   " + styGreen.Render("all done"))
	}
	return fit(b.String(), w)
}

func (m tuiModel) renderHelp() string {
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
		styKey.Render("q") + "           quit (asks for confirmation while running)",
		styKey.Render("ctrl+c") + "      quit immediately",
		styKey.Render("ctrl+l") + "      redraw the screen",
		"",
		styDim.Render("glyphs: · queued  ⟲ resolving  ⠋ downloading  ⚙ merging  ✓ done  ≡ duplicate  ✗ failed  ⇄ official video"),
		"",
		styDim.Render("full output is written to " + m.eng.LogPath()),
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

// --- drawing primitives ------------------------------------------------------

// renderBox draws a rounded box of exactly w x h cells with a title in the
// top border and an optional right aligned label.
func renderBox(title, right string, w, h int, lines []string, focused bool) string {
	bs := styBorder
	if focused {
		bs = styBorderOn
	}
	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 0 {
		innerH = 0
	}

	// top border: ╭─ Title ───── right ─╮
	tw := innerW - 4
	if right != "" {
		tw -= ansi.StringWidth(right) + 3
	}
	t := truncate(title, max(tw, 1))
	fill := innerW - 3 - ansi.StringWidth(t)
	if right != "" {
		fill -= ansi.StringWidth(right) + 3
	}
	if fill < 0 {
		fill = 0
	}
	top := bs.Render("╭─ ") + styTitle.Render(t) + bs.Render(" "+strings.Repeat("─", fill))
	if right != "" {
		top += bs.Render(" ") + styDim.Render(right) + bs.Render(" ─")
	}
	top += bs.Render("╮")

	var b strings.Builder
	b.WriteString(top)
	for i := 0; i < innerH; i++ {
		b.WriteString("\n")
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		b.WriteString(bs.Render("│") + padRight(truncate(line, innerW), innerW) + bs.Render("│"))
	}
	b.WriteString("\n" + bs.Render("╰"+strings.Repeat("─", innerW)+"╯"))
	return b.String()
}

// window returns the slice of rows that keeps index sel visible in h rows.
func window(rows []string, sel, h int) []string {
	if h <= 0 || len(rows) <= h {
		return rows
	}
	start := sel - h/2
	if start < 0 {
		start = 0
	}
	if start+h > len(rows) {
		start = len(rows) - h
	}
	return rows[start : start+h]
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= w {
		return s
	}
	if w == 1 {
		return ansi.Truncate(s, 1, "")
	}
	return ansi.Truncate(s, w, "…")
}

// wrap breaks s into lines of at most w cells on word boundaries.
func wrap(s string, w int) []string {
	if w <= 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for ansi.StringWidth(word) > w {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			lines = append(lines, ansi.Truncate(word, w, ""))
			word = word[len(ansi.Truncate(word, w, "")):]
		}
		switch {
		case line == "":
			line = word
		case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// fit pads or truncates s to exactly w cells.
func fit(s string, w int) string {
	return padRight(truncate(s, w), w)
}

func padRight(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// runTUI drives the full screen interface until the user quits.
func runTUI(ctx context.Context, eng *Engine) (*runState, error) {
	p := tea.NewProgram(newTUIModel(eng), tea.WithAltScreen(), tea.WithContext(ctx))
	final, err := p.Run()
	if fm, ok := final.(tuiModel); ok {
		return fm.state, err
	}
	return nil, err
}
