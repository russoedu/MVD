package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"context"
	"fmt"
	"os"
	"time"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/screenview"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

// Controller is what the screen needs from the engine.
type Controller = downloadsession.Controller

// Enabled decides between the full screen interface and plain log output.
func Enabled(noTUIFlag bool) bool {
	if noTUIFlag || os.Getenv("MVD_NO_TUI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Drive runs the download screen until the user quits and returns the final state for
// the summary.
func Drive(ctx context.Context, c Controller) (*runstate.State, error) {
	p := tea.NewProgram(New(c), tea.WithContext(ctx))
	final, err := p.Run()
	if fm, ok := final.(Model); ok {
		return fm.session.State, err
	}
	return nil, err
}

const (
	paneLists   = downloadsession.PaneLists
	paneEntries = downloadsession.PaneEntries
	paneOutput  = downloadsession.PaneOutput
)

// minWideWidth is the window width from which the output pane is shown beside the lists.
const minWideWidth = 100

// minPane is the narrowest a column can be dragged to.
const minPane = 20

type (
	evMsg    struct{ ev interface{} }
	evClosed struct{}
	tickMsg  time.Time
)

// Model is the download screen as a Bubble Tea model: the session holds what the screen does, and
// what is here is what only the terminal knows, such as sizes in cells, the dragging of borders and
// the widgets that edit text.
type Model struct {
	session downloadsession.Session
	events  <-chan interface{}

	width, height int
	// leftW and midW are the widths the person gave the columns by dragging their borders;
	// 0 is the default. dragging is the border being dragged: 1 after the playlists, 2 after
	// the entries, 0 for none. reserve is the rows kept under the playlists for buttons.
	leftW, dragging, reserve int
	// dragRow is the playlist being carried, dropRow the one it is over; dragging a row is on while
	// dragRowOn.
	dragRowOn        bool
	dragRow, dropRow int
	filterList       filterMenu
	// marquee counts the ticks, for the text of the selected row that is wider than its column.
	marquee int
	spin    int
	// embedded: inside an app model, it ends with FinishedMsg instead of quitting.
	embedded bool
	// playlistsTop and entriesTop are how many rows the links and the songs are scrolled down.
	playlistsTop, entriesTop int
	// colWidth are the widths the person gave to columns by carrying their borders. They are
	// replaced, never changed, so a copy of the screen keeps its own.
	colWidth   map[string]int
	colDrag    columnDrag
	columnList columnsMenu
	// swap and links are the widgets that edit the address of another video and the links to add;
	// the session holds what they say.
	swap  textinput.Model
	links textarea.Model
}

// FinishedMsg tells the app model the user left the download screen.
type FinishedMsg struct{}

// quit ends the screen: the program in RunDownload, or the screen alone when
// the model is part of an app model.
func (m Model) quit() tea.Cmd {
	if m.embedded {
		return func() tea.Msg { return FinishedMsg{} }
	}
	return tea.Quit
}

func New(c Controller) Model {
	return Model{session: downloadsession.New(c), events: c.Events()}
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

func (m Model) Init() tea.Cmd {
	if m.events == nil {
		return tick()
	}
	return tea.Batch(waitEvent(m.events), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	before := [3]int{m.session.SelPlaylist, m.session.SelEntry, int(m.session.Pane)}
	next, cmd := m.update(msg)
	if now, ok := next.(Model); ok && [3]int{now.session.SelPlaylist, now.session.SelEntry, int(now.session.Pane)} != before {
		// A row that was just chosen shows the start of its name, still, before it scrolls.
		now.marquee = 0
		now.keepSelectionVisible()
		next = now
	}

	return next, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.spin = (m.spin + 1) % len(look.SpinnerFrames)
		m.marquee++
		return m, tick()

	case evMsg:
		id := m.session.State.Apply(msg.ev)
		if added, ok := msg.ev.(engine.EvPlaylistAdded); ok {
			// A link that was just added is the one the person looks at: its songs come in under it.
			if at := m.session.State.PlaylistPos(added.Source.Index); at >= 0 {
				m.session.SelPlaylist, m.session.SelEntry, m.entriesTop = at, 0, 0
				m.keepSelectionVisible()
			}
		}
		m.session.SelPlaylist = max(min(m.session.SelPlaylist, len(m.session.State.Playlists)-1), 0)
		if !m.session.Loaded && (m.session.State.Idle || m.session.State.AllListed()) {
			m.session.Loaded = true
		}
		if st, ok := msg.ev.(engine.EvEntryState); ok && m.session.Follow {
			if st.State == engine.StateDownloading || st.State == engine.StateResolving {
				m.jumpTo(id)
			}
		}
		if lg, ok := msg.ev.(engine.EvLog); ok && m.session.OutScroll > 0 {
			// Keep the view anchored while the user is reading older lines.
			if en := m.selectedEntry(); en != nil && lg.Entry == en.ID {
				m.session.OutScroll++
			}
		}
		return m, waitEvent(m.events)

	case evClosed:
		if m.session.Hosted {
			// The run is over; what it reported stays on the screen.
			m.events, m.session.Ctrl = nil, nil
			return m, func() tea.Msg { return RunEndedMsg{} }
		}
		return m, m.quit()

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.mouse(msg)
	}
	if m.session.Adding {
		// The text area's cursor blinks on messages of its own.
		var cmd tea.Cmd
		m.links, cmd = m.links.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, m.quit()
	}
	if key == "ctrl+l" {
		return m, tea.ClearScreen
	}
	if m.session.Adding {
		return m.updateAdding(msg)
	}
	if m.session.Swapping {
		return m.updateSwapping(msg)
	}
	if m.columnList.open {
		return m.columnsKey(key), nil
	}
	if m.filterList.open {
		return m.filterKey(key), nil
	}
	if key == "c" && m.controlsShown() {
		m.columnList, m.filterList = columnsMenu{open: true}, filterMenu{}
		return m, nil
	}
	if key == "F" && m.controlsShown() {
		m.filterList, m.columnList = filterMenu{open: true}, columnsMenu{}
		return m, nil
	}

	if m.session.ConfirmQuit {
		switch key {
		case "y", "Y", "enter":
			return m, m.quit()
		default:
			m.session.ConfirmQuit = false
		}
		return m, nil
	}

	if m.session.ShowHelp {
		// A stray key does not close it; the ones that mean "done" do.
		switch key {
		case "esc", "?", "enter", "space", "q":
			m.session.ShowHelp = false
		}
		return m, nil
	}

	if next, cmd, handled := m.chooseKey(key); handled {
		return next, cmd
	}

	if m.session.Hosted && (key == "q" || key == "esc" || key == "a") {
		return m, nil // the workspace takes these
	}

	switch key {
	case "q", "esc":
		if m.session.State.Idle {
			return m, m.quit()
		}
		m.session.ConfirmQuit = true
	case "?":
		m.session.ShowHelp = true
	case "a":
		return m.startAdding()
	case "tab":
		m.session.Pane = (m.session.Pane + 1) % 3
		if m.session.Pane == paneOutput && !m.outputVisible() {
			m.session.Pane = paneLists
		}
	case "shift+tab":
		m.session.Pane = (m.session.Pane + 2) % 3
		if m.session.Pane == paneOutput && !m.outputVisible() {
			m.session.Pane = paneEntries
		}
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "shift+up", "shift+down":
		if _, held := m.chooser(); held && m.session.Pane == paneEntries {
			m.extendPick(map[string]int{"shift+up": -1, "shift+down": 1}[key])
		}
	case "pgup":
		m.move(-10)
	case "pgdown":
		m.move(10)
	case "home":
		m.move(-1 << 20)
	case "end":
		m.move(1 << 20)
	case "enter":
		m.session.Follow = !m.session.Follow
		if m.session.Follow {
			m.session.OutScroll = 0
			if id := m.activeEntry(); id >= 0 {
				m.jumpTo(id)
			}
			m.flash("following active download")
		} else {
			m.flash("follow off")
		}
	case "f":
		m.showOnly(songlist.GroupFailed)
		if m.filterActive() {
			m.flash("showing failed songs only")
		} else {
			m.flash("showing all songs")
		}
	case "t":
		m.session.Technical = !m.session.Technical
		m.session.OutScroll = 0
		if m.session.Technical {
			m.flash("technical log on: t goes back to the plain details")
		} else {
			m.flash("plain details on")
		}
	case "l":
		m.session.FullLog = !m.session.FullLog
		if m.session.FullLog {
			m.session.Technical = true // the full screen is for reading the log
			m.session.Pane = paneOutput
		} else if m.session.Pane == paneOutput && !m.outputVisible() {
			m.session.Pane = paneEntries
		}
	case "r":
		m.retry()
	}
	return m, nil
}

func (m *Model) flash(s string) {
	m.session.Notice = s
	m.session.NoticeUntil = time.Now().Add(3 * time.Second)
}

func (m *Model) outputVisible() bool {
	return m.session.FullLog || m.width >= minWideWidth
}

func (m *Model) move(delta int) {
	switch m.session.Pane {
	case paneLists:
		m.session.Follow = false
		m.session.SelPlaylist = viewport.Clamp(m.session.SelPlaylist+delta, 0, len(m.session.State.Playlists)-1)
		m.session.SelEntry = 0
		m.session.OutScroll = 0
	case paneEntries:
		m.session.Follow = false
		n := len(m.visibleEntries())
		m.session.SelEntry = viewport.Clamp(m.session.SelEntry+delta, 0, n-1)
		m.session.OutScroll = 0
	case paneOutput:
		lines := len(m.outputLines())
		m.session.OutScroll = viewport.Clamp(m.session.OutScroll-delta, 0, max(0, lines-1))
	}
}

func (m *Model) retry() {
	if m.session.Ctrl == nil {
		m.flash("nothing is running; start the downloads first")
		return
	}
	switch m.session.Pane {
	case paneEntries:
		en := m.selectedEntry()
		if en == nil {
			return
		}
		if en.State != engine.StateFailed {
			m.flash("only failed entries can be retried")
			return
		}
		if m.session.Ctrl.Retry(en.ID) {
			m.flash(fmt.Sprintf("retrying %02d %s", en.Index, entryTitle(en)))
		}
	default:
		if m.session.SelPlaylist < len(m.session.State.Playlists) {
			n := m.session.Ctrl.RetryPlaylist(m.session.State.Playlists[m.session.SelPlaylist].Index)
			m.flash(fmt.Sprintf("re-queued %d failed entries", n))
		}
	}
}

// visibleEntries returns the entry ids shown in the entries pane.
func (m *Model) visibleEntries() []int {
	if m.session.SelPlaylist >= len(m.session.State.Playlists) {
		return nil
	}
	pl := m.session.State.Playlists[m.session.SelPlaylist]
	ids := pl.Entries
	if m.filterActive() {
		ids = nil
		for _, id := range pl.Entries {
			if m.matchesFilter(m.session.State.Entries[id]) {
				ids = append(ids, id)
			}
		}
	}
	return m.sorted(ids)
}

func (m *Model) selectedEntry() *runstate.Entry {
	ids := m.visibleEntries()
	if len(ids) == 0 {
		return nil
	}
	i := viewport.Clamp(m.session.SelEntry, 0, len(ids)-1)
	return m.session.State.Entries[ids[i]]
}

// activeEntry returns the id of the first entry currently downloading.
func (m *Model) activeEntry() int {
	for _, en := range m.session.State.Entries {
		if en != nil && (en.State == engine.StateDownloading || en.State == engine.StateMerging || en.State == engine.StateResolving) {
			return en.ID
		}
	}
	return -1
}

func (m *Model) jumpTo(id int) {
	en := m.session.State.Entry(id)
	if en == nil {
		return
	}
	m.session.SelPlaylist = max(m.session.State.PlaylistPos(en.Playlist), 0)
	for i, eid := range m.visibleEntries() {
		if eid == id {
			m.session.SelEntry = i
			break
		}
	}
	m.session.OutScroll = 0
}

func (m Model) View() tea.View { return screenview.Of(m.render()) }

// NewEmbedded is the download screen as a part of an app model: it ends with FinishedMsg
// instead of quitting the program.
func NewEmbedded(c Controller) Model {
	m := New(c)
	m.embedded = true
	return m
}

// RunEndedMsg tells the workspace that the run behind a hosted screen has ended: its events
// are over, and the screen keeps showing what it reported.
type RunEndedMsg struct{}

// NewHosted is the download screen as the main part of the workspace: it begins with no run
// (showing whatever state it is given), takes a run with Attach and lets it go when it ends.
// Quitting and adding links are left to the workspace.
func NewHosted() Model {
	return Model{embedded: true, session: downloadsession.NewHosted()}
}

// WithState shows state, for a list that is not being run: what an earlier session
// reported, read from the list file.
func (m Model) WithState(state *runstate.State) Model {
	m.session.State = state
	m.session.SelPlaylist, m.session.SelEntry, m.session.OutScroll = 0, 0, 0
	return m
}

// Attach shows a run that has begun: the screen starts again from what the run reports. The
// command it returns waits for the run's first event.
func (m Model) Attach(c Controller) (Model, tea.Cmd) {
	m.session.Ctrl, m.events = c, c.Events()
	m.session.State = runstate.New(c.Sources())
	m.session.Loaded = len(m.session.State.Playlists) == 0
	m.session.Follow = true
	// The link the cursor was on stays the one, when the run has it: a link was just added there.
	m.session.SelPlaylist = max(min(m.session.SelPlaylist, len(m.session.State.Playlists)-1), 0)
	m.session.SelEntry, m.session.OutScroll, m.entriesTop, m.playlistsTop = 0, 0, 0, 0
	return m, waitEvent(m.events)
}

// Detach lets the run go, keeping what it reported on the screen.
func (m Model) Detach() Model {
	m.session.Ctrl, m.events, m.session.Paused = nil, nil, false
	return m
}

// WithPaused says whether the downloads are held, which the header shows.
func (m Model) WithPaused(paused bool) Model {
	m.session.Paused = paused
	return m
}

// Running reports whether a run is attached.
func (m Model) Running() bool { return m.session.Ctrl != nil }

// Resize gives the screen its size.
func (m Model) Resize(w, h int) Model {
	m.width, m.height = w, h
	return m
}

// IsLeftover is whether msg is one of the screen's own messages, left over from a run that
// has ended and no longer has a screen to take it.
func IsLeftover(msg tea.Msg) bool {
	switch msg.(type) {
	case tickMsg, evMsg, evClosed:
		return true
	}
	return false
}

// State is what the run has reported so far.
func (m Model) State() *runstate.State { return m.session.State }

// Render draws the screen, for an app model that shows it under its own notice.
func (m Model) Render() string { return m.render() }
