package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/sourcelist"
)

// SetupAction is what the user chose on the setup screens.
type SetupAction int

const (
	ActionQuit SetupAction = iota
	ActionStart
)

// SetupInput seeds the setup screens.
type SetupInput struct {
	Cfg        config.Config
	URLs       []string
	CfgPath    string
	ListPath   string
	OpenConfig bool // start on the config screen (first run)
	// PickFolder, when set, is the operating system's folder chooser, which the
	// folder settings open instead of the built-in folder browser.
	PickFolder FolderPicker
}

// SetupResult is returned when the setup screens close.
type SetupResult struct {
	Action SetupAction
	Cfg    config.Config
	URLs   []string
}

// RunSetup runs the list + config screens and returns the user's choice.
func RunSetup(in SetupInput) (SetupResult, error) {
	m := newSetupModel(in)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if fm, ok := final.(setupModel); ok {
		return fm.result, err
	}
	return SetupResult{Action: ActionQuit, Cfg: in.Cfg, URLs: in.URLs}, err
}

// NewSetupModel returns the list + config screens as a Bubble Tea model, for a
// host that runs the program itself (for example over the web). The returned
// model also has an Outline method (see ScreenOutline).
func NewSetupModel(in SetupInput) tea.Model { return newSetupModel(in) }

const (
	screenList = iota
	screenConfig
	screenAdvanced
)

type setupModel struct {
	screen        int
	width, height int
	cfg           config.Config
	cfgPath       string
	listPath      string
	list          listModel
	config        configModel
	advanced      advancedModel
	result        SetupResult
	embedded      bool // inside an app model: ends with setupFinishedMsg instead of quitting
	pick          FolderPicker
}

// setupFinishedMsg tells the app model the user left the setup screens.
type setupFinishedMsg struct{ result SetupResult }

// finish ends the screens: the program in RunSetup, or the screens alone when
// the model is part of an app model.
func (m setupModel) finish() tea.Cmd {
	if m.embedded {
		res := m.result
		return func() tea.Msg { return setupFinishedMsg{result: res} }
	}
	return tea.Quit
}

func newSetupModel(in SetupInput) setupModel {
	m := setupModel{
		cfg:      in.Cfg,
		cfgPath:  in.CfgPath,
		listPath: in.ListPath,
		list:     newListModel(in.URLs),
		config:   newConfigModel(in.Cfg).withFolderPicker(in.PickFolder),
		pick:     in.PickFolder,
		result:   SetupResult{Action: ActionQuit, Cfg: in.Cfg, URLs: in.URLs},
	}
	if in.OpenConfig {
		m.screen = screenConfig
	}
	return m
}

func (m setupModel) Init() tea.Cmd { return tea.Batch(textareaBlink, m.config.init()) }

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list = m.list.setSize(msg.Width, msg.Height)
		m.config = m.config.setSize(msg.Width, msg.Height)
		m.advanced = m.advanced.setSize(msg.Width, msg.Height)
		return m, nil
	case tea.MouseMsg:
		return m.mouse(msg)
	}

	switch m.screen {
	case screenList:
		next, cmd, out := m.list.update(msg)
		m.list = next
		switch out {
		case listStart:
			m.saveList()
			m.result = SetupResult{Action: ActionStart, Cfg: m.cfg, URLs: m.list.urls()}
			return m, m.finish()
		case listPrefs:
			m.saveList()
			m.config = newConfigModel(m.cfg).withFolderPicker(m.pick).setSize(m.width, m.height)
			m.screen = screenConfig
			return m, nil
		case listQuit:
			m.saveList()
			m.result = SetupResult{Action: ActionQuit, Cfg: m.cfg, URLs: m.list.urls()}
			return m, m.finish()
		}
		return m, cmd
	case screenConfig:
		next, cmd, out, cfg := m.config.update(msg)
		m.config = next
		switch out {
		case cfgSave:
			m.cfg = cfg
			_ = config.Save(m.cfg, m.cfgPath)
			m.screen = screenList
			return m, nil
		case cfgCancel:
			m.screen = screenList
			return m, nil
		case cfgAdvanced:
			m.advanced = newAdvancedModel(m.config.cfg).setSize(m.width, m.height)
			m.screen = screenAdvanced
			return m, nil
		}
		return m, cmd
	case screenAdvanced:
		next, cmd, out, cfg := m.advanced.update(msg)
		m.advanced = next
		switch out {
		case advSave:
			m.config.cfg = cfg
			m.screen = screenConfig
			return m, nil
		case advCancel:
			m.screen = screenConfig
			return m, nil
		}
		return m, cmd
	}
	return m, nil
}

func (m setupModel) View() string {
	switch m.screen {
	case screenConfig:
		return m.config.view(m.width, m.height)
	case screenAdvanced:
		return m.advanced.view(m.width, m.height)
	default:
		return m.list.view(m.width, m.height)
	}
}

func (m setupModel) saveList() {
	if m.listPath != "" {
		_ = sourcelist.Save(m.listPath, m.list.urls())
	}
}

// --- key bar shared by setup screens ----------------------------------------

type keyHint struct{ key, desc string }

func keyBar(width int, hints []keyHint) string {
	var b strings.Builder
	b.WriteString(" ")
	for i, h := range hints {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(styKey.Render(h.key) + " " + styDim.Render(h.desc))
	}
	return fit(b.String(), width)
}

func screenFrame(width, height int, title, body, bar string) string {
	head := styTitle.Render(" " + title)
	bodyH := height - 2
	if bodyH < 1 {
		bodyH = 1
	}
	body = lipgloss.NewStyle().Width(width).Height(bodyH).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, head, body, bar)
}
