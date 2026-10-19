package tui

import (
	"context"
	"reflect"
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/dialog"
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
		if a.inPrefs {
			a.prefs = a.prefs.sizedTo(a.popupSize())
		}
	case workspace.PreferencesRequestedMsg:
		in := a.in.Setup
		in.Cfg, in.OpenConfig = a.cfg, true
		a.prefs = newSetupModel(in)
		a.prefs.embedded, a.prefs.prefsOnly = true, true
		a.prefs, a.inPrefs = a.prefs.sizedTo(a.popupSize()), true
		return a, a.prefs.Init()
	case setupFinishedMsg:
		a.inPrefs = false
		if msg.result.Confirmed {
			a = a.apply(msg.result.Cfg, true)
		}
		ApplyTheme(a.cfg.Colors)
		return a, nil
	}
	if a.inPrefs {
		if _, ok := msg.(tea.MouseMsg); ok {
			return a, nil
		}
		next, cmd := a.prefs.Update(msg)
		a.prefs = next.(setupModel)
		// Each change is kept and goes to the run at once; the places the files go to and
		// the pattern of their names wait for the person to confirm with save.
		return a.apply(a.prefs.config.Config(), false), cmd
	}
	next, cmd := a.ws.Update(msg)
	a.ws = next.(workspace.Model)
	return a, cmd
}

// apply takes cfg as the settings, keeping the places and the name pattern as they were
// unless confirmed; it writes them to the config file and tells the run.
func (a workspaceApp) apply(cfg config.Config, confirmed bool) workspaceApp {
	if !confirmed {
		cfg.OutputDir, cfg.OutputTemplate, cfg.LogDir = a.cfg.OutputDir, a.cfg.OutputTemplate, a.cfg.LogDir
	}
	if reflect.DeepEqual(cfg, a.cfg) {
		return a
	}
	a.cfg = cfg
	_ = config.Save(cfg, a.in.Setup.CfgPath)
	live := engine.LiveSettings{}
	if a.in.Live != nil {
		live = a.in.Live(cfg)
	}
	a.ws = a.ws.Reconfigure(cfg, live)
	return a
}

// popupSize is the room the preferences take over the workspace.
func (a workspaceApp) popupSize() (int, int) {
	return max(min(a.width-6, 112), 1), max(min(a.height-4, 36), 1)
}

// sizedTo gives the screens the inside of a frame w by h.
func (m setupModel) sizedTo(w, h int) setupModel {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w - 2, Height: h - 2})
	return next.(setupModel)
}

func (a workspaceApp) View() tea.View {
	if !a.inPrefs {
		return a.ws.View()
	}
	w, h := a.popupSize()
	x, y := dialog.Centre(a.width, a.height, w, h)
	inner := strings.Split(a.prefs.render(), "\n")
	return screenview.Of(dialog.Overlay(a.ws.Render(), dialog.Frame("Preferences", inner, w, h), x, y))
}

// RunWorkspace runs the workspace app until the person quits it or ctx ends.
func RunWorkspace(ctx context.Context, in WorkspaceInput) error {
	_, err := tea.NewProgram(NewWorkspaceApp(in), tea.WithContext(ctx)).Run()
	return err
}
