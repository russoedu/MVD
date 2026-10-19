package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	editorpage "youtube-downloader/libs/mvd-core/editor/editor"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/menubar"
	"youtube-downloader/libs/mvd-core/playlistfile"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/screenview"
)

type (
	// runStartedMsg carries the run the host built, or why it could not.
	runStartedMsg struct {
		run    ListRun
		err    error
		review bool
	}
	// runClosedMsg says a run finished winding down.
	runClosedMsg struct{ quit bool }
	// filePickedMsg is what the operating system's file window answered.
	filePickedMsg struct {
		purpose purpose
		path    string
		chosen  bool
		err     error
	}
	// saveTickMsg is the beat at which the list is brought up to date with the run.
	saveTickMsg time.Time
)

// PreferencesRequestedMsg asks the host to show the preferences; it gives the result back
// with Reconfigure.
type PreferencesRequestedMsg struct{}

// saveEvery is how often the list file is brought up to date while a run goes on.
const saveEvery = 2 * time.Second

// purpose is what the one-line question at the bottom is asking for.
type purpose int

const (
	askNone purpose = iota
	askOpen
	askSaveAs
	askSaveThenStart
	askSaveThenReview
	askAddLink
)

// Model is the workspace.
type Model struct {
	host    Host
	cfg     config.Config
	session *downloadlist.Session

	bar    menubar.Model
	screen downloadpage.Model

	run      ListRun
	starting bool
	stopping bool

	editor    editorpage.Model
	reviewing bool
	decided   int // the version of the decisions of the review the list has

	edit    editPopup
	editing bool

	asking purpose
	input  textinput.Model

	notice        string
	width, height int
}

// New returns the workspace showing list, which may be nil for a new, empty one.
func New(host Host, cfg config.Config, list *downloadlist.Session) Model {
	if list == nil {
		list = downloadlist.New(nil)
	}
	bar := menubar.New(menubar.WithMenus(menus()...), menubar.WithTitle("MVD"))
	for _, name := range notYet {
		bar.SetEnabled(name, false)
	}
	m := Model{host: host, cfg: cfg, session: list, bar: bar, screen: downloadpage.NewHosted()}
	m.screen = m.screen.WithState(list.Preview())
	return m
}

// Session is the list that is open.
func (m Model) Session() *downloadlist.Session { return m.session }

func (m Model) Init() tea.Cmd { return m.screen.Init() }

func (m Model) running() bool { return m.run != nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.bar.SetSize(msg.Width, max(m.bar.ContentHeight(), 1))
		m.screen = m.screen.Resize(msg.Width, max(msg.Height-1, 0))
		next, _ := m.editor.Update(tea.WindowSizeMsg{Width: msg.Width, Height: max(msg.Height-1, 0)})
		m.editor = next
		return m, nil

	case menubar.SelectedMsg:
		return m.choose(msg.Name)

	case runStartedMsg:
		return m.runStarted(msg)

	case runClosedMsg:
		m.run, m.stopping, m.reviewing = nil, false, false
		m.screen = m.screen.Detach().WithState(m.session.Preview())
		if msg.quit {
			return m, m.quit()
		}
		return m, nil

	case saveTickMsg:
		if m.run == nil {
			return m, nil
		}
		m.syncList()
		return m, saveTick()

	case filePickedMsg:
		if msg.err != nil {
			m.notice = "Cannot show the file window: " + msg.err.Error()
			return m, nil
		}
		if !msg.chosen {
			return m, nil
		}
		return m.answered(msg.purpose, msg.path)

	case editSavedMsg:
		m.editing = false
		if err := m.session.Replace(msg.items); err != nil {
			m.notice = "Cannot save the list: " + err.Error()
		}
		m.screen = m.screen.WithState(m.session.Preview())
		return m, nil

	case editCancelledMsg:
		m.editing = false
		return m, nil

	case editorpage.LeaveMsg:
		m.reviewing = false
		return m, nil

	case downloadpage.RunEndedMsg:
		return m, nil
	}

	if editorpage.Owns(msg) {
		if !m.reviewing {
			return m, nil
		}
		return m.toEditor(msg)
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if m.asking != askNone {
			return m.answer(key)
		}
		if m.stopping || m.starting {
			return m, nil
		}
		if m.editing && m.bar.Open() < 0 {
			edit, cmd := m.edit.Update(key)
			m.edit = edit
			return m, cmd
		}
		if m.bar.Open() < 0 && !m.reviewing {
			if next, cmd, handled := m.runKey(key.String()); handled {
				return next, cmd
			}
		}
		if !m.reviewing && !m.editing || m.bar.Open() >= 0 {
			bar, cmd := m.bar.Update(msg)
			m.bar = bar
			if cmd != nil || m.bar.Open() >= 0 {
				return m, cmd
			}
		}
		if m.reviewing {
			return m.toEditor(msg)
		}
		return m.toScreen(msg)
	}

	if m.editing && m.bar.Open() < 0 {
		if click, ok := msg.(tea.MouseClickMsg); !ok || click.Y != 0 {
			return m, nil
		}
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && (click.Y == 0 || m.bar.Open() >= 0) {
		bar, cmd := m.bar.Update(msg)
		m.bar = bar
		return m, cmd
	}
	if m.bar.Open() >= 0 {
		bar, cmd := m.bar.Update(msg)
		m.bar = bar
		return m, cmd
	}
	if m.reviewing {
		return m.toEditor(msg)
	}
	return m.toScreen(msg)
}

// toScreen gives a message to the download screen, which sits one row under the top.
func (m Model) toScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	if click, ok := msg.(tea.MouseClickMsg); ok {
		click.Y--
		msg = click
	}
	next, cmd := m.screen.Update(msg)
	m.screen = next.(downloadpage.Model)
	return m, cmd
}

// toEditor gives a message to the review, and keeps the list up to date with what the
// person decided there.
func (m Model) toEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.editor.Update(msg)
	m.editor = next
	if m.editor.Version() != m.decided {
		m.decided = m.editor.Version()
		if err := m.session.Decide(m.editor.Rows()); err != nil {
			m.notice = "Cannot save the list: " + err.Error()
		}
	}
	return m, cmd
}

// runKey is what the keys of the workspace do while no menu or question is open.
func (m Model) runKey(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "s":
		next, cmd := m.start(false)
		return next, cmd, true
	case "g":
		next, cmd := m.start(true)
		return next, cmd, true
	case "p":
		return m.togglePause(), nil, true
	case "x":
		next, cmd := m.stop(false)
		return next, cmd, true
	case "R":
		return m.retryAll(), nil, true
	case "a":
		return m.ask(askAddLink, "Add a playlist or video link", ""), textinput.Blink, true
	case "e":
		if m.running() && m.editor.Live() {
			m.reviewing = true
			return m, nil, true
		}
		if !m.running() {
			return m.openEdit(), nil, true
		}
	}
	return m, nil, false
}

// choose does what a menu entry says.
func (m Model) choose(name string) (tea.Model, tea.Cmd) {
	if m.stopping || m.starting {
		return m, nil
	}
	switch name {
	case entryNew:
		if m.running() {
			m.notice = "Stop the downloads before starting a new list"
			return m, nil
		}
		m.session = downloadlist.New(nil)
		m.screen = m.screen.WithState(m.session.Preview())
		m.notice = ""
	case entryOpen:
		if m.running() {
			m.notice = "Stop the downloads before opening another list"
			return m, nil
		}
		if m.host.PickFile != nil {
			pick, dir := m.host.PickFile, m.host.ListDir
			return m, func() tea.Msg {
				path, chosen, err := pick(dir)
				return filePickedMsg{askOpen, path, chosen, err}
			}
		}
		return m.ask(askOpen, "Open the list", m.defaultPath()), textinput.Blink
	case entrySave:
		if m.session.Path() == "" {
			return m.askWhereToSave(askSaveAs)
		}
		m.save("")
	case entrySaveAs:
		return m.askWhereToSave(askSaveAs)
	case entryEditList:
		return m.openEdit(), nil
	case entryDelete, entryMoveUp, entryMoveDown, entryUndo, entryRedo:
		if !m.editing {
			m.notice = "Open Edit list first (Ctrl+E), then " + strings.ToLower(name) + " there"
			return m, nil
		}
		key := map[string]tea.KeyPressMsg{
			entryDelete:   {Code: tea.KeyDelete},
			entryMoveUp:   {Code: tea.KeyUp, Mod: tea.ModShift},
			entryMoveDown: {Code: tea.KeyDown, Mod: tea.ModShift},
			entryUndo:     {Code: 'z', Mod: tea.ModCtrl},
			entryRedo:     {Code: 'y', Mod: tea.ModCtrl},
		}[name]
		edit, cmd := m.edit.Update(key)
		m.edit = edit
		return m, cmd
	case entryPreferences:
		return m, func() tea.Msg { return PreferencesRequestedMsg{} }
	case entryAbout:
		m.notice = "MVD, Music Video Downloader"
	case entryOpenSource:
		if m.host.Open != nil {
			if err := m.host.Open(openSourceURL); err != nil {
				m.notice = "Cannot open the browser: " + err.Error()
			}
		}
	case entryQuit:
		return m.stop(true)
	}
	return m, nil
}

// openEdit shows the list to be edited. A list that is going cannot be, as the run has
// its own copy of it: it has to be stopped first.
func (m Model) openEdit() Model {
	if m.running() || m.starting {
		m.notice = "Stop the downloads before editing the list"
		return m
	}
	m.edit, m.editing = newEditPopup(m.session.File()), true
	return m
}

// askWhereToSave shows the file window for a list to be saved, or asks for the path on the
// bottom line when the host has none.
func (m Model) askWhereToSave(p purpose) (Model, tea.Cmd) {
	if m.host.SaveFile == nil {
		return m.ask(p, "Save the list as", m.defaultPath()), textinput.Blink
	}
	save, dir, name := m.host.SaveFile, filepath.Dir(m.defaultPath()), filepath.Base(m.defaultPath())
	return m, func() tea.Msg {
		path, chosen, err := save(dir, name)
		return filePickedMsg{p, path, chosen, err}
	}
}

func (m Model) defaultPath() string {
	if p := m.session.Path(); p != "" {
		return p
	}
	return filepath.Join(m.host.ListDir, "list"+playlistfile.Extension)
}

func (m *Model) save(path string) {
	if err := m.session.Save(path); err != nil {
		m.notice = "Cannot save the list: " + err.Error()
		return
	}
	m.notice = "Saved " + m.session.Path()
}

// ask opens the one-line question at the bottom of the screen.
func (m Model) ask(p purpose, prompt, value string) Model {
	in := textinput.New()
	in.Prompt = prompt + ": "
	in.SetValue(value)
	in.CursorEnd()
	in.Focus()
	m.asking, m.input = p, in
	return m
}

func (m Model) answer(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.asking = askNone
		return m, nil
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		p := m.asking
		m.asking = askNone
		if text == "" {
			return m, nil
		}
		return m.answered(p, text)
	}
	in, cmd := m.input.Update(key)
	m.input = in
	return m, cmd
}

func (m Model) answered(p purpose, text string) (tea.Model, tea.Cmd) {
	switch p {
	case askOpen:
		list, err := downloadlist.Open(withExtension(text))
		if err != nil {
			m.notice = "Cannot open the list: " + err.Error()
			return m, nil
		}
		m.session = list
		m.screen = m.screen.WithState(list.Preview())
		m.notice = "Opened " + list.Path()
	case askSaveAs:
		m.save(withExtension(text))
	case askSaveThenStart, askSaveThenReview:
		m.save(withExtension(text))
		if m.session.Path() == "" {
			return m, nil
		}
		return m.start(p == askSaveThenReview)
	case askAddLink:
		m.session.SetSources(append(m.session.Sources(), text))
		if adder, ok := m.run.(downloadpage.SourceAdder); ok && m.run != nil {
			adder.AddSource(text)
		} else {
			m.screen = m.screen.WithState(m.session.Preview())
		}
		m.syncFile()
	}
	return m, nil
}

func withExtension(path string) string {
	if filepath.Ext(path) == "" {
		return path + playlistfile.Extension
	}
	return path
}

// syncFile writes the list to its file, when it has one.
func (m *Model) syncFile() {
	if m.session.Path() == "" {
		return
	}
	if err := m.session.Save(""); err != nil {
		m.notice = "Cannot save the list: " + err.Error()
	}
}

// start begins the downloads of the list, or its review. A list with no file is saved
// first, so what happens to it is never lost.
func (m Model) start(review bool) (Model, tea.Cmd) {
	switch {
	case m.running():
		m.notice = "The downloads are already going"
		return m, nil
	case len(m.session.Sources()) == 0:
		m.notice = "The list is empty: add a link first (a)"
		return m, nil
	case m.host.Start == nil:
		m.notice = "Downloading is not available here"
		return m, nil
	}
	if m.session.Path() == "" {
		p := askSaveThenStart
		if review {
			p = askSaveThenReview
		}
		if m.host.SaveFile != nil {
			return m.askWhereToSave(p)
		}
		return m.ask(p, "Save the list as, then it starts", m.defaultPath()), textinput.Blink
	}
	if err := os.MkdirAll(m.cfg.OutputDir, 0o755); err != nil {
		m.notice = "Cannot create the download folder: " + err.Error()
		return m, nil
	}
	sources, plan := m.session.Inputs()
	m.starting, m.notice = true, ""
	host, cfg := m.host, m.cfg
	return m, func() tea.Msg {
		run, err := host.Start(cfg, sources, plan, review)
		return runStartedMsg{run: run, err: err, review: review}
	}
}

func (m Model) runStarted(msg runStartedMsg) (tea.Model, tea.Cmd) {
	m.starting = false
	if msg.err != nil {
		m.notice = "Cannot start the downloads: " + msg.err.Error()
		return m, nil
	}
	m.run = msg.run
	screen, cmd := m.screen.Attach(msg.run)
	m.screen = screen
	cmds := []tea.Cmd{cmd, saveTick()}
	if msg.review {
		m.editor = editorpage.NewLive(m.host.Editor, msg.run, m.session.File().Songs())
		m.decided = m.editor.Version()
		next, _ := m.editor.Update(tea.WindowSizeMsg{Width: m.width, Height: max(m.height-1, 0)})
		m.editor, m.reviewing = next, true
		cmds = append(cmds, m.editor.Init())
	}
	return m, tea.Batch(cmds...)
}

// syncList brings the list up to date with the run and writes it.
func (m *Model) syncList() {
	if m.run == nil {
		return
	}
	if err := m.session.SyncSnapshot(m.run.Snapshot()); err != nil {
		m.notice = "Cannot save the list: " + err.Error()
	}
}

func saveTick() tea.Cmd {
	return tea.Tick(saveEvery, func(t time.Time) tea.Msg { return saveTickMsg(t) })
}

// stop ends the run: the downloads under way are cancelled, and what each song got to is
// written to the list. With quit the app ends afterwards.
func (m Model) stop(quit bool) (Model, tea.Cmd) {
	if m.run == nil {
		if quit {
			return m, m.quit()
		}
		m.notice = "Nothing is running"
		return m, nil
	}
	m.syncList()
	m.stopping = true
	run, session := m.run, m.session
	return m, func() tea.Msg {
		run.Close()
		_ = session.SyncSnapshot(run.Snapshot())
		return runClosedMsg{quit: quit}
	}
}

func (m Model) quit() tea.Cmd {
	quit := m.host.Quit
	return func() tea.Msg {
		if quit != nil {
			quit()
		}
		return tea.Quit()
	}
}

func (m Model) togglePause() Model {
	if m.run == nil {
		m.notice = "Nothing is running"
		return m
	}
	if m.run.Paused() {
		m.run.Resume()
		m.notice = "Downloads go on"
	} else {
		m.run.Pause()
		m.notice = "Paused: the downloads under way finish, no new one starts"
	}
	return m
}

func (m Model) retryAll() Model {
	if m.run == nil {
		m.notice = "Nothing is running"
		return m
	}
	m.notice = fmt.Sprintf("%d songs queued again", m.run.RetryAll())
	return m
}

// Reconfigure tells the run, if there is one, that the settings changed.
func (m Model) Reconfigure(cfg config.Config, live engine.LiveSettings) Model {
	m.cfg = cfg
	if m.run != nil {
		m.run.Reconfigure(live)
	}
	return m
}

// View is the screen. While a menu is open the mouse is followed wherever it goes, so that
// moving over another heading opens its menu, as in a desktop app; otherwise only clicks and
// drags are reported.
func (m Model) View() tea.View {
	view := screenview.Of(m.render())
	if m.bar.Open() >= 0 {
		view.MouseMode = tea.MouseModeAllMotion
	}
	return view
}
