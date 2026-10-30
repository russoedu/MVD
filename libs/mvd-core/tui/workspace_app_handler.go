package tui

import (
	"context"
	"reflect"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/prefsform"
	"youtube-downloader/libs/mvd-core/prefspopup"
	"youtube-downloader/libs/mvd-core/screenview"
	"youtube-downloader/libs/mvd-core/workspace"
	"youtube-downloader/libs/mvd-core/workspace/windowsnapshot"
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
	// Style, when set, is called with the window style of the settings and its colour scheme when
	// the app starts and whenever either changes, and the preferences offer the choice. Window is then called with what
	// the window shows after every change while the style is a classic one, for the host to draw.
	Style  func(style, scheme string)
	Window func(windowsnapshot.Snapshot)
	// Preferences, when set, is called after every change with the preferences as a window draws
	// them while they are open, and nil once they close: the host draws them, in every style,
	// instead of the terminal look's popup.
	Preferences func(*prefsform.Window)
}

// NewWorkspaceApp returns the app whose main screen is the workspace: the download list
// under a menu bar. The preferences are shown over it from the File menu.
func NewWorkspaceApp(in WorkspaceInput) tea.Model {
	a := workspaceApp{in: in, ws: workspace.New(in.Host, in.Setup.Cfg, in.List), cfg: in.Setup.Cfg, hoverX: -1, hoverY: -1}
	if in.Setup.OpenConfig {
		a = a.openPrefs() // the first run: the settings have never been looked at
		a.prefs = a.prefs.WithWelcome()
	}
	return a
}

type workspaceApp struct {
	in            WorkspaceInput
	ws            workspace.Model
	cfg           config.Config
	prefs         prefspopup.Model
	inPrefs       bool
	// shortcutsOpen is whether the Keyboard shortcuts window is open over the preferences, and
	// shortcutsScroll how far down its lines it is.
	shortcutsOpen   bool
	shortcutsScroll int
	width, height int
	hoverX, hoverY int
	// modal is whether a popup was last reported open to Modal, and style the window style
	// last reported to Style.
	modal *bool
	style *string
}

// Init returns the first commands of the workspace, and of the preferences when they are open.
func (a workspaceApp) Init() tea.Cmd {
	if a.inPrefs {
		return tea.Batch(a.ws.Init(), a.prefs.Init())
	}
	return a.ws.Init()
}

// Update handles msg and returns the next model and its command. It tells the host when a popup opens or closes and when the window style changes.
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
	if app, ok := next.(workspaceApp); ok && app.in.Style != nil {
		next = app.report()
	}
	return next, cmd
}

// report tells the host the window style when it changed, and what the window shows while the
// style is a classic one; a screen only the terminal draws covers it, as the preferences do.
func (a workspaceApp) report() workspaceApp {
	style := a.cfg.WindowStyle
	if !config.ValidWindowStyle(style) {
		style = config.WindowStyleDOS
	}
	scheme := config.WindowSchemeOf(style, a.cfg.WindowScheme)
	if reported := style + "/" + scheme; a.style == nil || *a.style != reported {
		a.in.Style(style, scheme)
		a.style = &reported
	}
	if a.in.Preferences != nil {
		if a.inPrefs {
			w := a.prefs.Window()
			w.Pending = a.waitsForSave(a.prefs.Config())
			a.in.Preferences(&w)
		} else {
			a.in.Preferences(nil)
		}
	}
	if style == config.WindowStyleDOS || a.in.Window == nil {
		return a
	}
	snap := a.ws.Window()
	if a.inPrefs && a.in.Preferences == nil {
		snap.Covered = &windowsnapshot.Covered{Title: "Preferences"}
	}
	a.in.Window(snap)

	return a
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
	if key, ok := msg.(tea.KeyPressMsg); ok && a.inPrefs {
		switch {
		case a.shortcutsOpen:
			a.shortcutsScroll, a.shortcutsOpen = a.ws.ShortcutsKey(a.shortcutsScroll, key)

			return a, nil
		case key.String() == "f1":
			a.shortcutsOpen, a.shortcutsScroll = true, 0

			return a, nil
		}
	}
	if workspace.Background(msg) {
		next, cmd := a.ws.Update(msg) // the workspace's own beat goes on under a popup
		a.ws = next.(workspace.Model)
		return a, cmd
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
	chosen := a.prefs.Config()
	switch {
	case confirm:
		a = a.apply(chosen, true)
		a.ws = a.ws.Confirm("Preferences saved")
	case a.waitsForSave(chosen):
		a.ws = a.ws.Caution("The folder or file names you changed were not kept. Press s in Preferences to save them.")
	}
	a.inPrefs, a.shortcutsOpen = false, false
	ApplyTheme(a.cfg.Colors)
	return a
}

// waitsForSave says whether the preferences chosen differ from the ones in use in the settings
// that are only kept when the person saves: where the files go and how they are named.
func (a workspaceApp) waitsForSave(chosen config.Config) bool {
	return chosen.OutputDir != a.cfg.OutputDir ||
		chosen.OutputTemplate != a.cfg.OutputTemplate ||
		chosen.LogDir != a.cfg.LogDir
}

// popupHost is what the preferences ask of the machine: a folder chooser, and moving and
// removing the app.
func (a workspaceApp) popupHost() prefspopup.Host {
	in := a.in.Setup
	host := prefspopup.Host{StylesWindow: a.in.Style != nil}
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
	host.SignIn, host.SignInState = in.SignIn, in.SignInState
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

// View returns the workspace as a terminal view, with the preferences drawn over it when open.
func (a workspaceApp) View() tea.View {
	if !a.inPrefs {
		return a.ws.View()
	}
	if a.in.Preferences != nil {
		// The host draws the preferences as a window of its own, over the workspace.
		return screenview.Of(dialog.Dim(a.ws.Render()))
	}
	w, h := a.popupSize()
	x, y := dialog.Centre(a.width, a.height, w, h)
	screen := dialog.Overlay(dialog.Dim(a.ws.Render()), dialog.ClosableFrame("Preferences", a.popupLines(), w, h), x, y)
	if a.shortcutsOpen {
		view, sx, sy := a.ws.ShortcutsWindow(a.shortcutsScroll)
		screen = dialog.Overlay(dialog.Dim(screen), view, sx, sy)
	}

	return screenview.Of(screen)
}

// RunWorkspace runs the workspace app until the person quits it or ctx ends.
func RunWorkspace(ctx context.Context, in WorkspaceInput) error {
	_, err := tea.NewProgram(NewWorkspaceApp(in), tea.WithContext(ctx)).Run()
	return err
}
