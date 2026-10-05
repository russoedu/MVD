package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/sourcelist"
)

// NewAppModel returns the whole interactive app as one Bubble Tea model: the
// setup screens, then a download run when the user starts, then the setup
// screens again, the way the terminal app loops between RunSetup and
// RunDownload. A host that cannot run one program per screen (over the web, for
// instance) runs this model instead. It quits only when the user quits the
// setup screens. The returned model also has an Outline method (see
// ScreenOutline).
func NewAppModel(in AppInput) tea.Model {
	m := appModel{in: in, cfg: in.Setup.Cfg, urls: in.Setup.URLs}
	m.setup = m.newSetup(in.Setup.OpenConfig)
	return m
}

// runClosedMsg says a run finished winding down after the user left its screen.
type runClosedMsg struct{}

type appModel struct {
	in            AppInput
	cfg           config.Config
	urls          []string
	width, height int

	setup       setupModel
	download    model
	run         Run
	downloading bool
	stopping    bool   // the run is winding down; input is ignored until it has
	notice      string // why the user is back on the setup screens
}

func (m appModel) newSetup(openConfig bool) setupModel {
	in := m.in.Setup
	in.Cfg, in.URLs, in.OpenConfig = m.cfg, m.urls, openConfig
	s := newSetupModel(in)
	s.embedded = true
	if m.width == 0 {
		// The size is not known yet. Sizing the screens to nothing would wrap
		// the list into a sliver and leave it scrolled when the real size comes.
		return s
	}
	next, _ := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	return next.(setupModel)
}

func (m appModel) Init() tea.Cmd { return m.setup.Init() }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m.forward(msg)
	case setupFinishedMsg:
		return m.startRun(msg.result)
	case downloadFinishedMsg:
		return m.stopRun()
	case runClosedMsg:
		return m.backToSetup()
	}
	if m.stopping {
		return m, nil
	}
	if !m.downloading {
		switch msg.(type) {
		case tickMsg, evMsg, evClosed: // left over from a run that has ended
			return m, nil
		case tea.KeyMsg:
			m.notice = ""
		}
	}
	return m.forward(msg)
}

func (m appModel) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.downloading {
		next, cmd := m.download.Update(msg)
		m.download = next.(model)
		return m, cmd
	}
	next, cmd := m.setup.Update(msg)
	m.setup = next.(setupModel)
	return m, cmd
}

// startRun leaves the setup screens: it quits, stays when there is nothing to
// download, or starts a run and shows its screen.
func (m appModel) startRun(res SetupResult) (tea.Model, tea.Cmd) {
	m.cfg, m.urls = res.Cfg, res.URLs
	if res.Action == ActionQuit {
		return m, tea.Quit
	}
	if len(m.urls) == 0 {
		m.setup = m.newSetup(false)
		return m, m.setup.Init()
	}
	run, err := m.in.Start(m.cfg, m.urls)
	if err != nil {
		m.setup = m.newSetup(false)
		m.notice = "Cannot start the download: " + err.Error()
		return m, m.setup.Init()
	}
	dm := newModel(run)
	dm.embedded = true
	next, _ := dm.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.run, m.download, m.downloading = run, next.(model), true
	return m, m.download.Init()
}

// stopRun stops the run in the background, so the screen keeps drawing while
// the downloads wind down.
func (m appModel) stopRun() (tea.Model, tea.Cmd) {
	m.stopping = true
	run := m.run
	return m, func() tea.Msg {
		run.Close()
		return runClosedMsg{}
	}
}

// backToSetup returns to the setup screens once the run has closed, and
// empties the saved list when every item has been dealt with.
func (m appModel) backToSetup() (tea.Model, tea.Cmd) {
	if t := m.download.state.Tally(); t.Total > 0 && t.Queued == 0 && t.Running == 0 {
		_ = sourcelist.Clear(m.in.Setup.ListPath)
		m.urls = nil
	}
	m.run, m.downloading, m.stopping = nil, false, false
	m.setup = m.newSetup(false)
	return m, m.setup.Init()
}

func (m appModel) View() string {
	if m.downloading {
		return m.download.View()
	}
	view := m.setup.View()
	if m.notice == "" || m.width == 0 {
		return view
	}
	lines := strings.Split(view, "\n")
	lines[len(lines)-1] = fit(" "+styRed.Render(m.notice), m.width)
	return strings.Join(lines, "\n")
}
