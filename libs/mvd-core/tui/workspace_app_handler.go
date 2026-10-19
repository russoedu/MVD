package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/screenview"
	"youtube-downloader/libs/mvd-core/workspace"
)

// WorkspaceInput seeds the workspace app.
type WorkspaceInput struct {
	Setup SetupInput
	Host  workspace.Host
	// List is the list to open with, nil for a new one.
	List *downloadlist.Session
	// Live is the settings a run reads again at every download, taken from the config.
	Live func(config.Config) engine.LiveSettings
}

// NewWorkspaceApp returns the app whose main screen is the workspace: the download list
// under a menu bar. The preferences are shown over it from the File menu.
func NewWorkspaceApp(in WorkspaceInput) tea.Model {
	return workspaceApp{in: in, ws: workspace.New(in.Host, in.Setup.Cfg, in.List), cfg: in.Setup.Cfg}
}

type workspaceApp struct {
	in            WorkspaceInput
	ws            workspace.Model
	cfg           config.Config
	prefs         setupModel
	inPrefs       bool
	width, height int
}

func (a workspaceApp) Init() tea.Cmd { return a.ws.Init() }

func (a workspaceApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case workspace.PreferencesRequestedMsg:
		in := a.in.Setup
		in.Cfg, in.OpenConfig = a.cfg, true
		a.prefs = newSetupModel(in)
		a.prefs.embedded, a.prefs.prefsOnly = true, true
		next, _ := a.prefs.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.prefs, a.inPrefs = next.(setupModel), true
		return a, a.prefs.Init()
	case setupFinishedMsg:
		a.inPrefs = false
		a.cfg = msg.result.Cfg
		live := engine.LiveSettings{}
		if a.in.Live != nil {
			live = a.in.Live(a.cfg)
		}
		a.ws = a.ws.Reconfigure(a.cfg, live)
		return a, nil
	}
	if a.inPrefs {
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			next, cmd := a.prefs.Update(msg)
			a.prefs = next.(setupModel)
			return a, cmd
		}
		next, _ := a.prefs.Update(msg)
		a.prefs = next.(setupModel)
	}
	next, cmd := a.ws.Update(msg)
	a.ws = next.(workspace.Model)
	return a, cmd
}

func (a workspaceApp) View() tea.View {
	if a.inPrefs {
		return screenview.Of(a.prefs.render())
	}
	return a.ws.View()
}

// RunWorkspace runs the workspace app until the person quits it or ctx ends.
func RunWorkspace(ctx context.Context, in WorkspaceInput) error {
	_, err := tea.NewProgram(NewWorkspaceApp(in), tea.WithContext(ctx)).Run()
	return err
}
