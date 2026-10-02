package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"youtube-downloader/internal/engine"
	"youtube-downloader/internal/runstate"
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

// Run drives the full screen interface until the user quits and returns
// the final state for the summary.
func Run(ctx context.Context, c Controller) (*runstate.State, error) {
	p := tea.NewProgram(newModel(c), tea.WithAltScreen(), tea.WithContext(ctx))
	final, err := p.Run()
	if fm, ok := final.(model); ok {
		return fm.state, err
	}
	return nil, err
}

const (
	paneLists = iota
	paneEntries
	paneOutput
)

type (
	evMsg    struct{ ev interface{} }
	evClosed struct{}
	tickMsg  time.Time
)

type model struct {
	ctrl   Controller
	events <-chan interface{}
	state  *runstate.State

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

func newModel(c Controller) model {
	return model{
		ctrl:   c,
		events: c.Events(),
		state:  runstate.New(c.Sources()),
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

func (m model) Init() tea.Cmd {
	return tea.Batch(waitEvent(m.events), tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.spin = (m.spin + 1) % len(spinnerFrames)
		return m, tick()

	case evMsg:
		id := m.state.Apply(msg.ev)
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
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m *model) flash(s string) {
	m.status = s
	m.statusUntil = time.Now().Add(3 * time.Second)
}

func (m *model) outputVisible() bool {
	return m.fullLog || m.width >= minWideWidth
}

func (m *model) move(delta int) {
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

func (m *model) retry() {
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
func (m *model) visibleEntries() []int {
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

func (m *model) selectedEntry() *runstate.Entry {
	ids := m.visibleEntries()
	if len(ids) == 0 {
		return nil
	}
	i := clamp(m.selEntry, 0, len(ids)-1)
	return m.state.Entries[ids[i]]
}

// activeEntry returns the id of the first entry currently downloading.
func (m *model) activeEntry() int {
	for _, en := range m.state.Entries {
		if en != nil && (en.State == engine.StateDownloading || en.State == engine.StateMerging || en.State == engine.StateResolving) {
			return en.ID
		}
	}
	return -1
}

func (m *model) jumpTo(id int) {
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
