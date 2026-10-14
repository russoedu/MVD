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
	showHelp      bool
	confirmQuit   bool
	spin          int
	status        string // transient message shown in the key bar
	statusUntil   time.Time
	embedded      bool // inside an app model: ends with FinishedMsg instead of quitting
	adding        bool // the box for adding links is open
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
	return tea.Batch(waitEvent(m.events), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.spin = (m.spin + 1) % len(look.SpinnerFrames)
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
		m.showHelp = false
		return m, nil
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
