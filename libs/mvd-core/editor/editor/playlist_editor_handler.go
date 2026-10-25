package editor

import (
	"time"
	"youtube-downloader/libs/mvd-core/editor/review"
	"youtube-downloader/libs/mvd-core/look"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Messages of the editor. They have types of their own, so that the ones a plan run is
// still sending when the person has moved on never reach another screen.
type (
	EvMsg    struct{ Ev interface{} }
	EvClosed struct{}
	TickMsg  time.Time

	// PlannedMsg says the plan run is over: the app closes it.
	PlannedMsg struct{}
	// LeaveMsg says the person left the editor.
	LeaveMsg struct{}
	// DownloadMsg asks the app to download these songs; keys say which rows they are.
	DownloadMsg struct {
		Plan []engine.PlannedEntry
		Keys []string
	}
	// DescribedMsg carries the title of a video the person picked.
	DescribedMsg struct{ key, title string }
	// RefreshMsg is the beat at which a live editor looks at its run again.
	RefreshMsg time.Time
	// DecidedMsg says the person decided on a song of a live review, so the list that keeps
	// the decisions can be brought up to date.
	DecidedMsg struct{}
)

// Mode is what the editor is doing with the keys.
type Mode = review.Mode

const (
	Browsing    = review.Browsing
	AskPrevious = review.AskPrevious
	AskDownload = review.AskDownload
	Input       = review.Input
)

// Model is the playlist editor as a Bubble Tea model: the keys and the mouse become calls on a
// review.Session, which holds what the editor does; what is here is the screen's own: the window
// size, how far the list is scrolled, the widget of the one-line question and the spinner.
type Model struct {
	host    EditorHost
	session review.Session
	// events are the plan run's, while it works.
	events <-chan interface{}
	top    int

	width, height int

	spin int
	// input is the widget of the one-line question; asks is how many questions it has shown.
	input textinput.Model
	asks  int
}

// optionsOf says what the host is able to do.
func optionsOf(host EditorHost) review.Options {
	return review.Options{
		SessionFile: host.SessionFile,
		SaveDir:     host.SaveDir,
		DecisionLog: host.DecisionLog,
		CanOpen:     host.Open != nil,
		CanDescribe: host.Describe != nil,
		CanPickFile: host.PickFile != nil,
		CanSaveFile: host.SaveFile != nil,
		CanDownload: host.PlanDownload != nil,
	}
}

// NewFromRows opens the editor on rows that are already planned.
func NewFromRows(host EditorHost, rows []playlistfile.Entry) Model {
	return Model{host: host, session: review.NewFromRows(optionsOf(host), rows)}
}

// NewLive opens the editor on a run that holds its songs for review. brought are the
// decisions kept from an earlier session, which are put on the songs they belong to.
func NewLive(host EditorHost, live Live, brought []playlistfile.Entry) Model {
	return Model{host: host, session: review.NewLive(optionsOf(host), live, brought)}
}

// Live reports whether the editor shows a run as it goes.
func (m Model) Live() bool { return m.session.IsLive() }

// Version grows with every decision of a live review.
func (m Model) Version() int { return m.session.Version }

func refreshTick() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg { return RefreshMsg(t) })
}

// NewPlanning opens the editor while a plan-only run works. brought is what the
// person asked to bring from files; saved is the review of last time, which they are
// asked about when there is nothing else to bring.
func NewPlanning(host EditorHost, run downloadpage.Run, brought, saved []playlistfile.Entry) Model {
	return Model{
		host:    host,
		session: review.NewPlanning(optionsOf(host), run, brought, saved),
		events:  run.Events(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.session.IsLive() {
		return refreshTick()
	}
	if !m.session.Planning {
		return nil
	}
	return tea.Batch(waitEditorEvent(m.events), editorTick())
}

func waitEditorEvent(ch <-chan interface{}) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return EvClosed{}
		}
		return EvMsg{Ev: ev}
	}
}

func editorTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return TickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, m.width-24))
		return m, nil

	case TickMsg:
		if !m.session.Planning {
			return m, nil
		}
		m.spin = (m.spin + 1) % len(look.SpinnerFrames)
		return m, editorTick()

	case EvMsg:
		if !m.session.Planning {
			return m, nil
		}
		effects, over := m.session.ApplyEvent(msg.Ev)
		next, cmd := m.after(effects)
		if over {
			return next, cmd
		}
		return next, tea.Batch(cmd, waitEditorEvent(m.events))

	case EvClosed:
		return m, nil

	case RefreshMsg:
		if !m.session.IsLive() {
			return m, nil
		}
		m.session.Refresh()
		m.keepSelectionVisible()
		return m, refreshTick()

	case FilePickedMsg:
		m.session.FilePicked(msg.path, msg.chosen, msg.err, msg.save)
		return m.after(nil)

	case DescribedMsg:
		m.session.Described(msg.key, msg.title)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.mouse(msg)
	}

	if m.session.Mode == review.Input {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// after does what a call on the session asked of the host, shows a one-line question the call
// opened, and keeps the cursor on the screen.
func (m Model) after(effects []review.Effect) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, e := range effects {
		switch e := e.(type) {
		case review.Leave:
			cmds = append(cmds, func() tea.Msg { return LeaveMsg{} })
		case review.PlanFinished:
			cmds = append(cmds, func() tea.Msg { return PlannedMsg{} })
		case review.Download:
			plan, keys := e.Plan, e.Keys
			cmds = append(cmds, func() tea.Msg { return DownloadMsg{Plan: plan, Keys: keys} })
		case review.OpenURL:
			m.session.Opened(e.URL, m.host.Open(e.URL))
		case review.DescribeVideo:
			key, id, describe := e.Key, e.VideoID, m.host.Describe
			cmds = append(cmds, func() tea.Msg {
				title, err := describe(id)
				if err != nil {
					return nil
				}
				return DescribedMsg{key: key, title: title}
			})
		case review.PickSaveFile:
			pick, dir, name := m.host.SaveFile, e.Dir, e.Name
			cmds = append(cmds, func() tea.Msg {
				path, chosen, err := pick(dir, name)
				return FilePickedMsg{path: path, chosen: chosen, err: err, save: true}
			})
		case review.PickOpenFile:
			pick, start := m.host.PickFile, e.Start
			cmds = append(cmds, func() tea.Msg {
				path, chosen, err := pick(start)
				return FilePickedMsg{path: path, chosen: chosen, err: err}
			})
		}
	}
	if m.session.Asks != m.asks {
		m.asks = m.session.Asks
		in := textinput.New()
		in.Prompt = m.session.Prompt
		in.CharLimit = 0
		in.SetWidth(max(10, m.width-len(in.Prompt)-4))
		in.SetValue(m.session.InputText)
		in.CursorEnd()
		cmds = append(cmds, in.Focus())
		m.input = in
	}
	m.keepSelectionVisible()
	return m, tea.Batch(cmds...)
}

// Flash shows s on the bottom line for a moment.
func (m *Model) Flash(s string) { m.session.Flash(s) }

// WithDownloaded marks the songs with these keys as downloaded, so the next download
// skips them.
func (m Model) WithDownloaded(keys []string) Model {
	m.session.MarkDownloaded(keys)
	return m
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m.after([]review.Effect{review.Leave{}})
	}
	if key == "ctrl+l" {
		return m, tea.ClearScreen
	}
	if m.session.Picking {
		return m, nil // the file window is up
	}

	switch m.session.Mode {
	case review.AskPrevious:
		switch key {
		case "y", "Y", "enter":
			return m.after(m.session.AnswerPrevious(true))
		case "n", "N", "esc":
			return m.after(m.session.AnswerPrevious(false))
		}
		return m, nil
	case review.AskDownload:
		switch key {
		case "y", "Y", "enter":
			return m.after(m.session.AnswerDownload(review.DownloadReviewed))
		case "a", "A":
			return m.after(m.session.AnswerDownload(review.DownloadAsProposed))
		}
		return m.after(m.session.AnswerDownload(review.DownloadBack))
	case review.Input:
		return m.updateInput(msg)
	}

	if m.session.Planning {
		if key == "q" || key == "esc" {
			return m.after([]review.Effect{review.Leave{}})
		}
		return m, nil
	}

	s := &m.session
	switch key {
	case "q", "esc":
		return m.after([]review.Effect{review.Leave{}})
	case "up", "k":
		s.Move(-1)
	case "down", "j":
		s.Move(1)
	case "pgup":
		s.Move(-m.listHeight())
	case "pgdown":
		s.Move(m.listHeight())
	case "home":
		s.Move(-len(s.Rows))
	case "end":
		s.Move(len(s.Rows))
	case "o", "enter":
		return m.after(s.Open(false))
	case "v":
		return m.after(s.Open(true))
	case "c", " ":
		return m.after(s.Confirm())
	case "u":
		return m.after(s.Decide(playlistfile.DecisionOriginal, "", ""))
	case "x":
		return m.after(s.Decide(playlistfile.DecisionSkipped, "", ""))
	case "z", "backspace":
		return m.after(s.Decide(playlistfile.DecisionNone, "", ""))
	case "r":
		s.AskReplace()
	case "a":
		return m.after(s.ConfirmAll())
	case "d":
		return m.after(s.Download())
	case "s":
		return m.after(s.Save())
	case "l":
		return m.after(s.Load())
	default:
		return m, nil
	}
	return m.after(nil)
}

func (m *Model) keepSelectionVisible() {
	h := m.listHeight()
	sel := m.session.Selected
	if sel < m.top {
		m.top = sel
	}
	if sel >= m.top+h {
		m.top = sel - h + 1
	}
	m.top = viewport.Clamp(m.top, 0, max(0, len(m.session.Rows)-h))
}

// updateInput handles a key while the one-line question is open: enter answers it, esc leaves it,
// anything else is typed into it.
func (m Model) updateInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.session.CancelInput()
		return m, nil
	case "enter":
		m.session.SetInput(m.input.Value())
		return m.after(m.session.SubmitInput())
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.session.SetInput(m.input.Value())
	return m, cmd
}

// FilePickedMsg is the answer of the system's open-file window.
type FilePickedMsg struct {
	path   string
	chosen bool
	err    error
	// save is true for the answer of the save-file window.
	save bool
}

// Owns is whether msg is one of the editor's own messages, which an app model forwards to
// it while it is on the screen.
func Owns(msg tea.Msg) bool {
	switch msg.(type) {
	case EvMsg, EvClosed, TickMsg, DescribedMsg, FilePickedMsg, RefreshMsg:
		return true
	}
	return false
}

// Mode is what the editor is doing with the keys.
func (m Model) Mode() Mode { return m.session.Mode }

// Planning is whether a plan-only run is still working.
func (m Model) Planning() bool { return m.session.Planning }

// Rows are the songs, with what was proposed and decided for each.
func (m Model) Rows() []playlistfile.Entry { return m.session.Rows }

// Selected is the index of the song the cursor is on.
func (m Model) Selected() int { return m.session.Selected }

// WithInput puts text in the one-line question that is open.
func (m Model) WithInput(text string) Model {
	m.input.SetValue(text)
	m.session.SetInput(text)
	return m
}
