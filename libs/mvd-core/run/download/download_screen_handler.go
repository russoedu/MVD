package download

import (
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
)

// Controller is what the screen needs from the engine.
type Controller interface {
	Events() <-chan interface{}
	Sources() []engine.PlaylistSource
	LogPath() string
	Retry(entryID int) bool
	RetryPlaylist(playlist int) int
}

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
		return fm.state, err
	}
	return nil, err
}

const (
	paneLists = iota
	paneEntries
	paneOutput
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

type Model struct {
	ctrl   Controller
	events <-chan interface{}
	state  *runstate.State

	width, height int
	focus         int
	selPlaylist   int
	selEntry      int
	outScroll     int // lines scrolled up from the bottom of the output pane
	follow        bool
	loaded        bool // every playlist has been listed; until then the loading screen is shown
	filterFailed  bool
	fullLog       bool
	// technical shows the log of the tool in the details pane instead of the plain account.
	technical bool
	// paused is whether the downloads are held: the header says so for as long as they are.
	paused bool
	// leftW and midW are the widths the person gave the columns by dragging their borders;
	// 0 is the default. dragging is the border being dragged: 1 after the playlists, 2 after
	// the entries, 0 for none. reserve is the rows kept under the playlists for buttons.
	leftW, dragging, reserve int
	// editable is true while the list can be changed here: each playlist then has a grip to
	// carry it by, and icons to change and to remove it. dragRow is the playlist being
	// carried, dropRow the one it is over; dragging a row is on while dragRowOn.
	editable          bool
	dragRowOn         bool
	dragRow, dropRow  int
	// marquee counts the ticks, for the text of the selected row that is wider than its column.
	marquee int
	showHelp      bool
	confirmQuit   bool
	spin          int
	status        string // transient message shown in the key bar
	statusUntil   time.Time
	embedded      bool // inside an app model: ends with FinishedMsg instead of quitting
	hosted        bool // inside the workspace: it owns quitting and adding links, and the run comes and goes
	adding        bool // the box for adding links is open
	// swapping is the address of another video being asked for the song swapEntry.
	swapping  bool
	swap      textinput.Model
	swapEntry int
	links         textarea.Model
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
	state := runstate.New(c.Sources())
	return Model{
		ctrl:   c,
		events: c.Events(),
		state:  state,
		follow: true,
		// A run with no playlists to list (the download of a plan) has nothing to wait for.
		loaded: len(state.Playlists) == 0,
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

func (m Model) Init() tea.Cmd {
	if m.events == nil {
		return tick()
	}
	return tea.Batch(waitEvent(m.events), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.spin = (m.spin + 1) % len(look.SpinnerFrames)
		m.marquee++
		return m, tick()

	case evMsg:
		id := m.state.Apply(msg.ev)
		if !m.loaded && (m.state.Idle || m.state.AllListed()) {
			m.loaded = true
		}
		if st, ok := msg.ev.(engine.EvEntryState); ok && m.follow {
			if st.State == engine.StateDownloading || st.State == engine.StateResolving {
				m.jumpTo(id)
			}
		}
		if lg, ok := msg.ev.(engine.EvLog); ok && m.outScroll > 0 {
			// Keep the view anchored while the user is reading older lines.
			if en := m.selectedEntry(); en != nil && lg.Entry == en.ID {
				m.outScroll++
			}
		}
		return m, waitEvent(m.events)

	case evClosed:
		if m.hosted {
			// The run is over; what it reported stays on the screen.
			m.events, m.ctrl = nil, nil
			return m, func() tea.Msg { return RunEndedMsg{} }
		}
		return m, m.quit()

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.mouse(msg)
	}
	if m.adding {
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
	if m.adding {
		return m.updateAdding(msg)
	}
	if m.swapping {
		return m.updateSwapping(msg)
	}

	if m.confirmQuit {
		switch key {
		case "y", "Y", "enter":
			return m, m.quit()
		default:
			m.confirmQuit = false
		}
		return m, nil
	}

	if m.showHelp {
		// A stray key does not close it; the ones that mean "done" do.
		switch key {
		case "esc", "?", "enter", "space", "q":
			m.showHelp = false
		}
		return m, nil
	}

	if next, cmd, handled := m.chooseKey(key); handled {
		return next, cmd
	}

	if m.hosted && (key == "q" || key == "esc" || key == "a") {
		return m, nil // the workspace takes these
	}

	switch key {
	case "q", "esc":
		if m.state.Idle {
			return m, m.quit()
		}
		m.confirmQuit = true
	case "?":
		m.showHelp = true
	case "a":
		return m.startAdding()
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
			m.flash("showing failed songs only")
		} else {
			m.flash("showing all songs")
		}
	case "t":
		m.technical = !m.technical
		m.outScroll = 0
		if m.technical {
			m.flash("technical log on: t goes back to the plain details")
		} else {
			m.flash("plain details on")
		}
	case "l":
		m.fullLog = !m.fullLog
		if m.fullLog {
			m.technical = true // the full screen is for reading the log
			m.focus = paneOutput
		} else if m.focus == paneOutput && !m.outputVisible() {
			m.focus = paneEntries
		}
	case "r":
		m.retry()
	}
	return m, nil
}

func (m *Model) flash(s string) {
	m.status = s
	m.statusUntil = time.Now().Add(3 * time.Second)
}

func (m *Model) outputVisible() bool {
	return m.fullLog || m.width >= minWideWidth
}

func (m *Model) move(delta int) {
	switch m.focus {
	case paneLists:
		m.follow = false
		m.selPlaylist = viewport.Clamp(m.selPlaylist+delta, 0, len(m.state.Playlists)-1)
		m.selEntry = 0
		m.outScroll = 0
	case paneEntries:
		m.follow = false
		n := len(m.visibleEntries())
		m.selEntry = viewport.Clamp(m.selEntry+delta, 0, n-1)
		m.outScroll = 0
	case paneOutput:
		lines := len(m.outputLines())
		m.outScroll = viewport.Clamp(m.outScroll-delta, 0, max(0, lines-1))
	}
}

func (m *Model) retry() {
	if m.ctrl == nil {
		m.flash("nothing is running; start the downloads first")
		return
	}
	switch m.focus {
	case paneEntries:
		en := m.selectedEntry()
		if en == nil {
			return
		}
		if en.State != engine.StateFailed {
			m.flash("only failed entries can be retried")
			return
		}
		if m.ctrl.Retry(en.ID) {
			m.flash(fmt.Sprintf("retrying %02d %s", en.Index, entryTitle(en)))
		}
	default:
		if m.selPlaylist < len(m.state.Playlists) {
			n := m.ctrl.RetryPlaylist(m.selPlaylist)
			m.flash(fmt.Sprintf("re-queued %d failed entries", n))
		}
	}
}

// visibleEntries returns the entry ids shown in the entries pane.
func (m *Model) visibleEntries() []int {
	if m.selPlaylist >= len(m.state.Playlists) {
		return nil
	}
	pl := m.state.Playlists[m.selPlaylist]
	if !m.filterFailed {
		return pl.Entries
	}
	var out []int
	for _, id := range pl.Entries {
		if m.state.Entries[id].State == engine.StateFailed {
			out = append(out, id)
		}
	}
	return out
}

func (m *Model) selectedEntry() *runstate.Entry {
	ids := m.visibleEntries()
	if len(ids) == 0 {
		return nil
	}
	i := viewport.Clamp(m.selEntry, 0, len(ids)-1)
	return m.state.Entries[ids[i]]
}

// activeEntry returns the id of the first entry currently downloading.
func (m *Model) activeEntry() int {
	for _, en := range m.state.Entries {
		if en != nil && (en.State == engine.StateDownloading || en.State == engine.StateMerging || en.State == engine.StateResolving) {
			return en.ID
		}
	}
	return -1
}

func (m *Model) jumpTo(id int) {
	en := m.state.Entry(id)
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
	return Model{embedded: true, hosted: true, loaded: true, follow: true, state: runstate.New(nil)}
}

// WithState shows state, for a list that is not being run: what an earlier session
// reported, read from the list file.
func (m Model) WithState(state *runstate.State) Model {
	m.state = state
	m.selPlaylist, m.selEntry, m.outScroll = 0, 0, 0
	return m
}

// Attach shows a run that has begun: the screen starts again from what the run reports. The
// command it returns waits for the run's first event.
func (m Model) Attach(c Controller) (Model, tea.Cmd) {
	m.ctrl, m.events = c, c.Events()
	m.state = runstate.New(c.Sources())
	m.loaded = len(m.state.Playlists) == 0
	m.follow = true
	m.selPlaylist, m.selEntry, m.outScroll = 0, 0, 0
	return m, waitEvent(m.events)
}

// Detach lets the run go, keeping what it reported on the screen.
func (m Model) Detach() Model {
	m.ctrl, m.events, m.paused = nil, nil, false
	return m
}

// WithPaused says whether the downloads are held, which the header shows.
func (m Model) WithPaused(paused bool) Model {
	m.paused = paused
	return m
}

// Running reports whether a run is attached.
func (m Model) Running() bool { return m.ctrl != nil }

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
func (m Model) State() *runstate.State { return m.state }

// Render draws the screen, for an app model that shows it under its own notice.
func (m Model) Render() string { return m.render() }
