package tui

import (
	"charm.land/bubbles/v2/textarea"
	"path/filepath"
	"youtube-downloader/libs/mvd-core/screenview"
	advancedpage "youtube-downloader/libs/mvd-core/setup/advanced"
	colourspage "youtube-downloader/libs/mvd-core/setup/colours"
	listpage "youtube-downloader/libs/mvd-core/setup/list"
	preferencespage "youtube-downloader/libs/mvd-core/setup/preferences"
	songspage "youtube-downloader/libs/mvd-core/setup/songs"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/sourcelist"
)

// textareaBlink drives the textarea cursor.
var textareaBlink = textarea.Blink

// SetupAction is what the user chose on the setup screens.
type SetupAction int

const (
	ActionQuit SetupAction = iota
	ActionStart
	// ActionReview plans the list and lets the person check it before downloading.
	ActionReview
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
	// Uninstall, when set, is offered on the preferences: it asks the person on the
	// machine's own screen and removes the app if they agree.
	Uninstall Uninstaller
	// Move, when set, is offered on the preferences: it asks the person on the machine's
	// own screen where the app should live and moves it there if they agree.
	Move Mover
	// SignIn, when set, is offered on the preferences: it opens a window to sign in to YouTube
	// and says what came of it.
	SignIn func() (message string, err error)
	// SignInState says what the last sign-in came to, for the preferences to show.
	SignInState func() string
}

// SetupResult is returned when the setup screens close.
type SetupResult struct {
	// Confirmed is true when the preferences were left with save, not cancelled.
	Confirmed bool
	Action    SetupAction
	Cfg       config.Config
	URLs      []string
}

// RunSetup runs the list + config screens and returns the user's choice.
func RunSetup(in SetupInput) (SetupResult, error) {
	m := newSetupModel(in)
	p := tea.NewProgram(m)
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
	screenColours
	screenSongs
)

type setupModel struct {
	screen        int
	width, height int
	cfg           config.Config
	cfgPath       string
	listPath      string
	list          listpage.Model
	config        preferencespage.Model
	advanced      advancedpage.Model
	colours       colourspage.Model
	songs         songspage.Model
	result        SetupResult
	prefsOnly     bool // only the preferences: leaving them ends the screens
	embedded      bool // inside an app model: ends with setupFinishedMsg instead of quitting
	pick          FolderPicker
	uninstall     Uninstaller
	move          Mover
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
		cfg:       in.Cfg,
		cfgPath:   in.CfgPath,
		listPath:  in.ListPath,
		list:      listpage.New(in.URLs),
		songs:     songspage.New(""),
		config:    preferencespage.New(in.Cfg).WithFolderPicker(in.PickFolder).WithUninstaller(in.Uninstall).WithMover(in.Move),
		pick:      in.PickFolder,
		uninstall: in.Uninstall,
		move:      in.Move,
		result:    SetupResult{Action: ActionQuit, Cfg: in.Cfg, URLs: in.URLs},
	}
	if in.OpenConfig {
		m.screen = screenConfig
	}
	return m
}

// Init returns the commands that make the text cursors blink.
func (m setupModel) Init() tea.Cmd { return tea.Batch(textareaBlink, m.config.Init()) }

// Update handles msg and returns the next model and its command. It sends msg to the screen shown and finishes the setup when that screen asks.
func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list = m.list.SetSize(msg.Width, msg.Height)
		m.config = m.config.SetSize(msg.Width, msg.Height)
		m.advanced = m.advanced.SetSize(msg.Width, msg.Height)
		m.colours = m.colours.SetSize(msg.Width, msg.Height)
		m.songs = m.songs.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.MouseMsg:
		return m.mouse(msg)
	}

	switch m.screen {
	case screenList:
		next, cmd, out := m.list.Update(msg)
		m.list = next
		switch out {
		case listpage.Start:
			m.saveList()
			m.result = SetupResult{Action: ActionStart, Cfg: m.cfg, URLs: m.list.URLs()}
			return m, m.finish()
		case listpage.Review:
			m.saveList()
			m.result = SetupResult{Action: ActionReview, Cfg: m.cfg, URLs: m.list.URLs()}
			return m, m.finish()
		case listpage.Prefs:
			m.saveList()
			m.config = preferencespage.New(m.cfg).WithFolderPicker(m.pick).WithUninstaller(m.uninstall).WithMover(m.move).SetSize(m.width, m.height)
			m.screen = screenConfig
			return m, nil
		case listpage.Songs:
			if m.listPath == "" {
				return m, nil
			}
			m.songs = songspage.New(filepath.Join(filepath.Dir(m.listPath), "songs")).SetSize(m.width, m.height)
			m.screen = screenSongs
			return m, textareaBlink
		case listpage.Quit:
			m.saveList()
			m.result = SetupResult{Action: ActionQuit, Cfg: m.cfg, URLs: m.list.URLs()}
			return m, m.finish()
		}
		return m, cmd
	case screenConfig:
		next, cmd, out, cfg := m.config.Update(msg)
		m.config = next
		switch out {
		case preferencespage.Save:
			m.cfg = cfg
			_ = config.Save(m.cfg, m.cfgPath)
			m.screen = screenList
			if m.prefsOnly {
				m.result = SetupResult{Confirmed: true, Action: ActionQuit, Cfg: m.cfg, URLs: m.list.URLs()}
				return m, m.finish()
			}
			return m, nil
		case preferencespage.Cancel:
			ApplyTheme(m.cfg.Colors)
			m.screen = screenList
			if m.prefsOnly {
				m.result = SetupResult{Action: ActionQuit, Cfg: m.cfg, URLs: m.list.URLs()}
				return m, m.finish()
			}
			return m, nil
		case preferencespage.Advanced:
			m.advanced = advancedpage.New(m.config.Config()).SetSize(m.width, m.height)
			m.screen = screenAdvanced
			return m, nil
		case preferencespage.Colours:
			m.colours = colourspage.New(m.config.Config()).SetSize(m.width, m.height)
			m.screen = screenColours
			return m, nil
		}
		return m, cmd
	case screenAdvanced:
		next, cmd, out, cfg := m.advanced.Update(msg)
		m.advanced = next
		switch out {
		case advancedpage.Save:
			m.config = m.config.WithConfig(cfg)
			m.screen = screenConfig
			return m, nil
		case advancedpage.Cancel:
			m.screen = screenConfig
			return m, nil
		}
		return m, cmd
	case screenSongs:
		next, cmd, out := m.songs.Update(msg)
		m.songs = next
		switch out {
		case songspage.Added:
			m.list = m.list.WithURL(m.songs.Saved())
			m.saveList()
			m.screen = screenList
			return m, nil
		case songspage.Cancel:
			m.screen = screenList
			return m, nil
		}
		return m, cmd
	case screenColours:
		next, cmd, out, cfg := m.colours.Update(msg)
		m.colours = next
		switch out {
		case colourspage.Save:
			m.config = m.config.WithConfig(cfg)
			m.screen = screenConfig
			return m, nil
		case colourspage.Cancel:
			ApplyTheme(m.config.Config().Colors)
			m.screen = screenConfig
			return m, nil
		}
		return m, cmd
	}
	return m, nil
}

func (m setupModel) render() string {
	switch m.screen {
	case screenConfig:
		return m.config.View(m.width, m.height)
	case screenAdvanced:
		return m.advanced.View(m.width, m.height)
	case screenColours:
		return m.colours.View(m.width, m.height)
	case screenSongs:
		return m.songs.View(m.width, m.height)
	default:
		return m.list.View(m.width, m.height)
	}
}

func (m setupModel) saveList() {
	if m.listPath != "" {
		_ = sourcelist.Save(m.listPath, m.list.URLs())
	}
}

// View returns the current setup screen as a terminal view.
func (m setupModel) View() tea.View { return screenview.Of(m.render()) }
