package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"context"
	"os"
	"time"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/screenview"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
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
	// evMsg and evClosed say which run's events they came from, so what a run that was let go
	// still says is not taken for what the run now shown says.
	evMsg struct {
		ev   interface{}
		from <-chan interface{}
	}
	evClosed struct{ from <-chan interface{} }
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
	// barDrag is the scrollbar being dragged.
	barDrag barKind
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
			return evClosed{from: ch}
		}
		return evMsg{ev: ev, from: ch}
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
		if msg.from != nil && msg.from != m.events {
			return m, nil // a run that has been let go: its events are not the ones shown
		}
		m.session.Apply(msg.ev)
		if added, ok := msg.ev.(engine.EvPlaylistAdded); ok && m.session.State.PlaylistPos(added.Source.Index) >= 0 {
			// A link that was just added is the one the person looks at: its songs come in under it.
			m.entriesTop = 0
			m.keepSelectionVisible()
		}
		return m, waitEvent(m.events)

	case evClosed:
		if msg.from != nil && msg.from != m.events {
			return m, nil
		}
		eff :=m.session.RunClosed()
		if _, ended := eff.(downloadsession.RunEnded); ended {
			m.events = nil
		}
		return m, m.run(eff)

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

// run turns what the session asks of the host into the command that does it.
func (m Model) run(eff downloadsession.Effect) tea.Cmd {
	switch e := eff.(type) {
	case downloadsession.Quit:
		return m.quit()
	case downloadsession.OpenURL:
		return func() tea.Msg { return OpenVideoMsg{URL: e.URL} }
	case downloadsession.RunEnded:
		return func() tea.Msg { return RunEndedMsg{} }
	}

	return nil
}

// handleKey turns a key into the intent it stands for in the part of the screen that has the keys.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, m.run(m.session.LeaveNow())
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
	if key == "c" && m.session.ControlsShown() {
		m.columnList, m.filterList = columnsMenu{open: true}, filterMenu{}
		return m, nil
	}
	if key == "F" && m.session.ControlsShown() {
		m.filterList, m.columnList = filterMenu{open: true}, columnsMenu{}
		return m, nil
	}

	if m.session.ConfirmQuit {
		return m, m.run(m.session.AnswerQuit(key == "y" || key == "Y" || key == "enter"))
	}

	if m.session.ShowHelp {
		// A stray key does not close it; the ones that mean "done" do.
		switch key {
		case "esc", "?", "enter", "space", "q":
			m.session.CloseHelp()
		}
		return m, nil
	}

	if next, cmd, handled := m.chooseKey(key); handled {
		return next, cmd
	}

	switch key {
	case "q", "esc":
		return m, m.run(m.session.RequestQuit())
	case "?":
		m.session.OpenHelp()
	case "a":
		return m.startAdding()
	case "tab":
		m.session.NextPane(m.outputVisible())
	case "shift+tab":
		m.session.PreviousPane(m.outputVisible())
	case "up", "k":
		m.session.Move(-1)
	case "down", "j":
		m.session.Move(1)
	case "shift+up":
		m.session.ExtendPick(-1)
	case "shift+down":
		m.session.ExtendPick(1)
	case "pgup":
		m.session.Move(-10)
	case "pgdown":
		m.session.Move(10)
	case "home":
		m.session.Move(-1 << 20)
	case "end":
		m.session.Move(1 << 20)
	case "enter":
		m.session.ToggleFollow()
	case "f":
		m.session.ToggleFailedOnly()
		m.entriesTop = 0
	case "t":
		m.session.ToggleTechnical()
	case "l":
		m.session.ToggleFullLog(m.width >= minWideWidth)
	case "r":
		m.session.Retry()
	}
	return m, nil
}

func (m *Model) outputVisible() bool {
	return m.session.FullLog || m.width >= minWideWidth
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
	m.session.Show(state)
	return m
}

// Attach shows a run that has begun: the screen starts again from what the run reports. The
// command it returns waits for the run's first event.
func (m Model) Attach(c Controller) (Model, tea.Cmd) {
	m.session.Attach(c)
	m.events = c.Events()
	m.entriesTop, m.playlistsTop = 0, 0
	return m, waitEvent(m.events)
}

// Detach lets the run go, keeping what it reported on the screen.
func (m Model) Detach() Model {
	m.session.Detach()
	m.events = nil
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

// Session is what the screen shows, for a window drawn by something other than the terminal.
func (m Model) Session() downloadsession.Session { return m.session }

// ChoosePlaylist puts the cursor on playlist i, as a click on it does.
func (m Model) ChoosePlaylist(i int) Model {
	m.session.ChoosePlaylist(i)

	return m
}

// ChooseEntry puts the cursor on the song at row i of the songs shown, as a click on it does.
func (m Model) ChooseEntry(i int) Model {
	m.session.ChooseEntry(i)

	return m
}

// SetTicked ticks or unticks the song id, as a click on its box does.
func (m Model) SetTicked(id int, on bool) Model {
	if m.session.Picked[id] != on {
		m.session.TickSong(id, false)
	}

	return m
}

// Render draws the screen, for an app model that shows it under its own notice.
func (m Model) Render() string { return m.render() }
