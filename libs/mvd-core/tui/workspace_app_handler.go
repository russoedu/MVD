package tui

import (
	"context"
	"reflect"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/prefspopup"
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
	// Modal, when set, is called with true when a popup opens over the screen and false when
	// the last one closes, so a host can dim what is around the terminal.
	Modal func(open bool)
}

// NewWorkspaceApp returns the app whose main screen is the workspace: the download list
// under a menu bar. The preferences are shown over it from the File menu.
func NewWorkspaceApp(in WorkspaceInput) tea.Model {
	a := workspaceApp{in: in, ws: workspace.New(in.Host, in.Setup.Cfg, in.List), cfg: in.Setup.Cfg, hoverX: -1, hoverY: -1}
	if in.Setup.OpenConfig {
		a = a.openPrefs() // the first run: the settings have never been looked at
	}
	return a
}

type workspaceApp struct {
	in            WorkspaceInput
	ws            workspace.Model
	cfg           config.Config
	prefs         prefspopup.Model
	inPrefs       bool
	width, height int
	hoverX, hoverY int
	// modal is whether a popup was last reported open to Modal.
	modal *bool
}

func (a workspaceApp) Init() tea.Cmd {
	if a.inPrefs {
		return tea.Batch(a.ws.Init(), a.prefs.Init())
	}
	return a.ws.Init()
}

func (a workspaceApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := a.update(msg)
	if app, ok := next.(workspaceApp); ok && app.in.Modal != nil {
		open := app.inPrefs || app.ws.WindowOpen()
		if app.modal == nil || *app.modal != open {
			app.in.Modal(open)
			app.modal = &open
		}
		next = app
	}
	return next, cmd
}

func (a workspaceApp) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		// The workspace behind the preferences needs the size too: on the first start the
		// preferences are open at once, and it would never be drawn.
		next, _ := a.ws.Update(msg)
		a.ws = next.(workspace.Model)
		if a.inPrefs {
			a.prefs = a.prefs.SetSize(a.bodySize())
		}
	case workspace.PreferencesRequestedMsg:
		a = a.openPrefs()
		return a, a.prefs.Init()
	case prefspopup.CloseMsg:
		return a.closePrefs(false), nil
	case prefspopup.SaveMsg:
		return a.closePrefs(true), nil
	}
	if mouse, ok := msg.(tea.MouseMsg); ok {
		a.hoverX, a.hoverY = mouse.Mouse().X, mouse.Mouse().Y
	}
	if a.inPrefs {
		if mouse, ok := msg.(tea.MouseMsg); ok {
			if a.onClose(mouse) {
				return a.closePrefs(false), nil
			}
			inner, inside := a.inPopup(mouse)
			if !inside {
				return a, nil
			}
			var forward bool
			if a, msg, forward = a.popupMouse(inner); !forward {
				return a, nil
			}
		}
		var cmd tea.Cmd
		a.prefs, cmd = a.prefs.Update(msg)
		// Each change is kept and goes to the run at once; the places the files go to and
		// the pattern of their names wait for the person to confirm with save.
		return a.apply(a.prefs.Config(), false), cmd
	}
	next, cmd := a.ws.Update(msg)
	a.ws = next.(workspace.Model)
	return a, cmd
}

// openPrefs shows the preferences over the workspace.
func (a workspaceApp) openPrefs() workspaceApp {
	a.prefs = prefspopup.New(a.cfg, a.popupHost()).SetSize(a.bodySize())
	a.inPrefs = true
	return a
}

// closePrefs takes the preferences away; with confirm, the places the files go to and the
// pattern of their names are kept too.
func (a workspaceApp) closePrefs(confirm bool) workspaceApp {
	if confirm {
		a = a.apply(a.prefs.Config(), true)
	}
	a.inPrefs = false
	ApplyTheme(a.cfg.Colors)
	return a
}

// popupHost is what the preferences ask of the machine: a folder chooser, and moving and
// removing the app.
func (a workspaceApp) popupHost() prefspopup.Host {
	in := a.in.Setup
	host := prefspopup.Host{}
	if in.PickFolder != nil {
		host.PickFolder = in.PickFolder
	}
	if in.Move != nil {
		move := in.Move
		host.Move = func() (string, error) {
			outcome, err := move()
			return movingResult(outcome, err)
		}
	}
	if in.Uninstall != nil {
		uninstall := in.Uninstall
		host.Uninstall = func() (string, error) {
			outcome, err := uninstall()
			return removalResult(outcome, err)
		}
	}
	return host
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
	return dialog.Sized(a.width, a.height, popupPercent, 60, 20)
}

// popupPercent is how much of the screen the preferences take, in each direction.
const popupPercent = 85

// Outline describes the current screen (see ScreenOutline).
func (a workspaceApp) Outline() ScreenOutline {
	if a.inPrefs {
		return a.prefs.Outline()
	}
	return a.ws.Outline()
}

func (a workspaceApp) View() tea.View {
	if !a.inPrefs {
		return a.ws.View()
	}
	w, h := a.popupSize()
	x, y := dialog.Centre(a.width, a.height, w, h)
	return screenview.Of(dialog.Overlay(dialog.Dim(a.ws.Render()), dialog.ClosableFrame("Preferences", a.popupLines(), w, h), x, y))
}

// RunWorkspace runs the workspace app until the person quits it or ctx ends.
func RunWorkspace(ctx context.Context, in WorkspaceInput) error {
	_, err := tea.NewProgram(NewWorkspaceApp(in), tea.WithContext(ctx)).Run()
	return err
}
