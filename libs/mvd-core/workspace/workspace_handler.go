package workspace

import (
	"youtube-downloader/libs/mvd-core/workspace/listhistory"
	"youtube-downloader/libs/mvd-core/workspace/closequestion"
	"youtube-downloader/libs/mvd-core/workspace/runcontrol"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/menubar"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/pointer"
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
	// toolsTickMsg is the beat at which the state of the tools is looked at again, while they
	// are being fetched.
	toolsTickMsg time.Time
)

// PreferencesRequestedMsg asks the host to show the preferences; it gives the result back
// with Reconfigure.
type PreferencesRequestedMsg struct{}

// OpenListMsg asks the workspace to open the list at Path, as when a .mvd file is double
// clicked while the app is already open.
type OpenListMsg struct{ Path string }

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
	askChangeLink
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
	// afterStop is what the person asked for while a run only held the list: it is done once the run
	// has been stopped.
	afterStop tea.Msg
	// loadedKey is the list of addresses the last run was started on (or taken in), so a list is
	// loaded in the background only when it has changed since.
	loadedKey string

	about bool
	// closing is the question asked before the app closes while songs download.
	closing closequestion.State
	// shortcuts is whether the Keyboard shortcuts window is showing, and shortcutsScroll how far
	// down its lines it is.
	shortcuts       bool
	shortcutsScroll int
	// inline is the address being written in its row; clicked is the last click, for double clicks.
	inline  inlineEdit
	clicked lastClick
	// hoverX and hoverY are where the pointer last was, for the buttons to light up.
	hoverX, hoverY int

	// history is the lists as they were before the changes made to it, for undo and redo.
	history listhistory.History
	// changing is the playlist whose address the question at the bottom is asking for.
	changing int

	asking purpose
	input  textinput.Model

	notice string
	// done is what an action that worked says, and caution what it warns of; both go away at the
	// next key or click, while an error stays until something replaces it.
	done, caution string
	// wasFinished is whether the run had finished when the screen was last drawn.
	wasFinished   bool
	width, height int
}

// New returns the workspace showing list, which may be nil for a new, empty one.
func New(host Host, cfg config.Config, list *downloadlist.Session) Model {
	if list == nil {
		list = downloadlist.New(nil)
	}
	bar := menubar.New(menubar.WithMenus(menus()...), menubar.WithTitle("MVD"))
	m := Model{hoverX: -1, hoverY: -1, host: host, cfg: cfg, session: list, bar: bar, screen: downloadpage.NewHosted()}
	m.screen = m.screen.WithState(list.Preview())
	return m
}

// Session is the list that is open.
func (m Model) Session() *downloadlist.Session { return m.session }

func (m Model) Init() tea.Cmd { return tea.Batch(m.screen.Init(), m.toolsTick()) }

func (m Model) running() bool { return m.run != nil }

// sourcesKey is the addresses of the list, as one text that changes when the list does.
func (m Model) sourcesKey() string { return strings.Join(m.session.Sources(), "\n") }

// wantsLoading says whether the songs of the list are to be loaded now, in the background: the list
// has addresses nobody has loaded, nothing is being written or asked, and the tools are there.
func (m Model) wantsLoading() bool {
	if !m.host.AutoLoad || m.host.Start == nil || m.running() || m.starting || m.stopping {
		return false
	}
	if m.inline.on || m.asking != askNone || m.about || m.shortcuts || len(m.session.Sources()) == 0 {
		return false
	}
	if _, waiting := m.toolsWaiting(); waiting {
		return false
	}

	return m.sourcesKey() != m.loadedKey
}

// facts are what the run is doing, for the decisions of package runcontrol.
func (m Model) facts() runcontrol.Facts {
	f := runcontrol.Facts{HasRun: m.run != nil, Starting: m.starting, Stopping: m.stopping}
	if m.run != nil {
		f.Paused = m.run.Paused()
		_, f.CanAddLinks = m.run.(downloadpage.SourceAdder)
	}
	if state := m.screen.State(); state != nil {
		f.Idle, f.AllListed, f.Tally = state.Idle, state.AllListed(), state.Tally()
	}

	return f
}

// addable says whether links can be added to the run that is going on.
func (m Model) addable() bool { return m.facts().Addable() }

// addToRun gives the run more links, and keeps them in the list.
func (m Model) addToRun(links []string) Model {
	adder, ok := m.run.(downloadpage.SourceAdder)
	if !ok || len(links) == 0 {
		return m
	}
	m.session.SetSources(append(m.session.Sources(), links...))
	for _, link := range links {
		adder.AddSource(link)
	}
	m.loadedKey = m.sourcesKey()
	m.syncFile()
	if len(links) == 1 {
		m.done = "Added: its songs are loading"
	} else {
		m.done = fmt.Sprintf("Added %d links: their songs are loading", len(links))
	}

	return m
}

// waiting says whether some song of the run waits for the person to choose it.
func (m Model) waiting() bool { return m.facts().Waiting() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	if ws, ok := next.(Model); ok {
		// When a run ends, what was said while it began (a draft was kept...) gives way to how it went.
		if finished := ws.finished(); finished != ws.wasFinished {
			ws.wasFinished = finished
			if finished {
				ws.done, ws.caution = "", ""
			}
		}
		if ws.wantsLoading() {
			var load tea.Cmd
			ws, load = ws.start(true)
			cmd = tea.Batch(cmd, load)
		}
		next = ws.refreshMenus().refreshRunMenus().keepRoomForButtons()
	}
	return next, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg, tea.MouseClickMsg:
		m.done, m.caution, m.notice = "", "", ""
	}
	if mouse, ok := msg.(tea.MouseMsg); ok {
		m.hoverX, m.hoverY = mouse.Mouse().X, mouse.Mouse().Y
		if _, moved := msg.(tea.MouseMotionMsg); moved && m.bar.Open() < 0 {
			m.bar, _ = m.bar.Update(msg) // the heading under the pointer lights up
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.bar.SetSize(msg.Width, max(m.bar.ContentHeight(), 1))
		m.screen = m.screen.Resize(msg.Width, max(msg.Height-topRows, 0))
		return m, nil

	case menubar.SelectedMsg:
		return m.choose(msg.Name)

	case runStartedMsg:
		return m.runStarted(msg)

	case runClosedMsg:
		left := m.screen.State().Tally()
		m.run, m.stopping = nil, false
		m.screen = m.screen.Detach().WithState(m.session.Preview())
		if unfinished := left.Queued + left.Running; unfinished > 0 && !msg.quit {
			m.done = fmt.Sprintf("Stopped: %d saved, %d not downloaded. Load the playlists again to carry on.", left.Done, unfinished)
		}
		if msg.quit {
			return m, m.quit()
		}
		if next := m.afterStop; next != nil {
			m.afterStop = nil
			return m.Update(next)
		}
		return m, nil

	case toolsTickMsg:
		return m, m.toolsTick()
	case saveTickMsg:
		if m.run == nil {
			return m, nil
		}
		m.syncList()
		return m, saveTick()

	case OpenListMsg:
		m = m.settle()
		if m.starting || m.stopping {
			m.notice = "One moment..."
			return m, nil
		}
		if m.running() {
			if m.holdsOnlyTheList() {
				return m.stopThen(msg)
			}
			m.notice = "Stop the downloads before opening " + filepath.Base(msg.Path)
			return m, nil
		}
		m.asking = askNone
		return m.answered(askOpen, msg.Path)

	case filePickedMsg:
		if msg.err != nil {
			m.notice = "Cannot show the file window: " + msg.err.Error()
			return m, nil
		}
		if !msg.chosen {
			return m, nil
		}
		return m.answered(msg.purpose, msg.path)

	case downloadpage.ListEditMsg:
		return m.editList(msg)

	case tea.ClipboardMsg:
		return m.pasted(msg.Content), nil

	case tea.PasteMsg:
		if m.inline.on {
			return m.pasteInline(msg.Content), nil
		}
		if m.asking == askNone {
			return m.addAddresses(msg.Content), nil
		}

	case downloadpage.OpenVideoMsg:
		if m.host.Open != nil {
			if err := m.host.Open(msg.URL); err != nil {
				m.notice = "Cannot open the browser: " + err.Error()
			}
		}
		return m, nil

	case downloadpage.RunEndedMsg:
		return m, nil

	case CloseRequestedMsg:
		return m.closeRequested()
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		if m.closing.On {
			return m.closeKey(key)
		}
		if m.about {
			if name := key.String(); name == "esc" || name == "enter" || name == "space" || name == "q" {
				m.about = false
			}
			return m, nil
		}
		if m.shortcuts {
			return m.shortcutsKey(key), nil
		}
		if m.asking != askNone {
			return m.answer(key)
		}
		if m.stopping || m.starting {
			return m, nil
		}
		if m.inline.on && m.bar.Open() < 0 {
			return m.inlineKey(key)
		}
		bar, cmd := m.bar.Update(msg)
		m.bar = bar
		if cmd != nil || m.bar.Open() >= 0 {
			return m, cmd
		}
		return m.toScreen(msg)
	}

	if mouse, ok := msg.(tea.MouseMsg); ok && m.bar.Open() < 0 {
		if next, cmd, handled := m.popupMouse(mouse); handled {
			return next, cmd
		}
		next, cmd, handled := m.listMouse(mouse)
		if handled {
			return next, cmd
		}
		m = next
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && m.bar.Open() < 0 && click.Button == tea.MouseLeft {
		if next, cmd, handled := m.toolbarClick(click.X, click.Y); handled {
			return next, cmd
		}
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && (click.Y == 0 || m.bar.Open() >= 0) {
		bar, cmd := m.bar.Update(msg)
		m.bar = bar
		return m, cmd
	}
	if downloadpage.IsLeftover(msg) {
		// The screen's own events and ticks go to it whatever is open: if one is taken by the menu
		// the screen is never asked for the next, and the run's events pile up until it stops.
		return m.toScreen(msg)
	}
	if m.bar.Open() >= 0 {
		bar, cmd := m.bar.Update(msg)
		m.bar = bar
		return m, cmd
	}
	return m.toScreen(msg)
}

// toScreen gives a message to the download screen, which sits under the menu bar and the toolbar.
func (m Model) toScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouse, ok := msg.(tea.MouseMsg); ok {
		msg = pointer.Shift(mouse, 0, -topRows)
	}
	next, cmd := m.screen.Update(msg)
	m.screen = next.(downloadpage.Model)
	return m, cmd
}

// choose does what a menu entry says.
func (m Model) choose(name string) (tea.Model, tea.Cmd) {
	if m.stopping || m.starting {
		return m, nil
	}
	if settles(name) {
		m = m.settle() // a run that has finished is over: nothing has to be stopped
	}
	if next, cmd, handled := m.runChoice(name); handled {
		return next, cmd
	}
	switch name {
	case entryNew:
		m = m.stopInline(false)
		if m.running() {
			if m.holdsOnlyTheList() {
				return m.stopThen(menubar.SelectedMsg{Name: name})
			}
			m.notice = "Stop the downloads before starting a new list"
			return m, nil
		}
		m.session = downloadlist.New(nil)
		m.screen = m.screen.WithState(m.session.Preview())
		m.notice = ""
	case entryOpen:
		if m.running() {
			if m.holdsOnlyTheList() {
				return m.stopThen(menubar.SelectedMsg{Name: name})
			}
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
	case entryUndo, entryRedo, entryCut, entryCopy, entryPaste, entryDelete, entryMoveUp, entryMoveDown:
		return m.editCommand(name)
	case entryPreferences:
		return m, func() tea.Msg { return PreferencesRequestedMsg{} }
	case entryAbout:
		m.about = true
	case entryOpenFolder:
		m.openFolder()
	case entryShortcuts:
		m.shortcuts, m.shortcutsScroll = true, 0
	case entryDocs:
		m.openAddress(projectURL + "#readme")
	case entryReport:
		m.openAddress(m.reportAddress())
	case entryQuit:
		return m.closeRequested()
	}
	return m, nil
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
	m.done = "Saved " + cells.TruncateMiddle(m.session.Path(), max(m.width-14, 20))
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
		m.done = "Opened " + cells.TruncateMiddle(list.Path(), max(m.width-14, 20))
	case askSaveAs:
		m.save(withExtension(text))
	case askSaveThenStart, askSaveThenReview:
		m.save(withExtension(text))
		if m.session.Path() == "" {
			return m, nil
		}
		return m.start(p == askSaveThenReview)
	case askAddLink, askChangeLink:
		if reason := notALink(text); reason != "" {
			m.notice = reason
			prompt := "Add a playlist or video link"
			if p == askChangeLink {
				prompt = "Change the link"
			}

			return m.ask(p, prompt, text), textinput.Blink // asked again, with what was typed
		}
		links, _ := parseLinks(text)
		return m.answeredLink(p, strings.Join(links, "\n"))
	}
	return m, nil
}

// answeredLink does what a link written on the bottom line says: it adds the link to the
// run, or to the list, or changes the one being changed.
func (m Model) answeredLink(p purpose, text string) (tea.Model, tea.Cmd) {
	switch p {
	case askAddLink:
		if m.addable() {
			m = m.addToRun([]string{text})
		} else {
			m = m.addAddresses(text)
		}
	case askChangeLink:
		m = m.changeAddress(m.changing, text)
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
		m.notice = "The list is empty: add a link first (Ctrl+L)"
		return m, nil
	case m.host.Start == nil:
		m.notice = "Downloading is not available here"
		return m, nil
	}
	if state, waiting := m.toolsWaiting(); waiting {
		m.caution, m.done = state.Message, ""
		return m, nil
	}
	m.loadedKey = m.sourcesKey()
	if m.session.Path() == "" {
		// A new list is kept as a draft in the app's own folder, so a download starts at once and
		// nothing is lost; File > Save as gives it a name of the person's own.
		if draft := m.draftPath(); draft != "" {
			if err := m.session.Save(draft); err != nil {
				m.notice = "Cannot save the list: " + err.Error()
				return m, nil
			}
			m.done = "Kept this list as a draft: " + cells.TruncateMiddle(draft, max(m.width-34, 20))
		}
	}
	if m.session.Path() == "" {
		p := askSaveThenReview
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
	m.screen = screen.WithPaused(msg.run.Paused())
	cmds := []tea.Cmd{cmd, saveTick()}
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
		m.done = "Downloading: what you accept goes on from here"
	} else {
		m.run.Pause()
		m.done = "Stopped: the songs on their way finish, and nothing else begins"
	}
	m.screen = m.screen.WithPaused(m.run.Paused())

	return m
}

func (m Model) retryAll() Model {
	if m.run == nil {
		m.notice = "Nothing is running"
		return m
	}
	m.done = fmt.Sprintf("%d songs queued again", m.run.RetryAll())
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
	view.MouseMode = tea.MouseModeAllMotion // the buttons light up under the pointer
	return view
}

// finished says whether the run has nothing more to do (see runcontrol).
func (m Model) finished() bool { return m.facts().Finished() }

// active says whether a run is going on that has to be stopped before the list can change: one
// that has finished does not.
func (m Model) active() bool { return m.facts().Active() }

// settle ends a run that has finished, so the list can be changed or another run started: what
// each song got to is written to the list, the run is closed (it has nothing to wind down, so this
// is quick) and the screen goes back to the list. A run that has not finished is left alone.
func (m Model) settle() Model {
	if !m.finished() {
		return m
	}
	m.syncList()
	m.run.Close()
	m.run = nil
	m.screen = m.screen.Detach().WithState(m.session.Preview())

	return m
}

// holdsOnlyTheList says whether the run does nothing that stopping it would lose (see runcontrol).
func (m Model) holdsOnlyTheList() bool { return m.facts().HoldsOnlyTheList() }

// stopThen stops the run, and does what msg asks once it has wound down.
func (m Model) stopThen(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.afterStop = msg

	return m.stop(false)
}

// settles says whether an entry needs the list to be free, and so ends a finished run first.
func settles(name string) bool {
	switch name {
	case entryNew, entryOpen, entryStart, entryAddLink,
		entryUndo, entryRedo, entryCut, entryPaste, entryDelete, entryMoveUp, entryMoveDown:
		return true
	}

	return false
}

// openFolder shows the folder of the videos in the system's file manager.
func (m *Model) openFolder() {
	dir := m.cfg.OutputDir
	switch {
	case dir == "":
		m.notice = "No folder is chosen for the videos yet: choose one in File > Preferences"
	case m.host.OpenFolder == nil:
		m.caution = "The videos are in " + cells.TruncateMiddle(dir, max(m.width-24, 20))
	default:
		if _, err := os.Stat(dir); err != nil {
			m.caution = "The folder is made when the first download starts: " + cells.TruncateMiddle(dir, max(m.width-60, 20))

			return
		}
		if err := m.host.OpenFolder(dir); err != nil {
			m.notice = "Cannot open the folder: " + err.Error()
		}
	}
}

// draftPath is where a list that has no file yet is kept when it starts: "Untitled list.mvd" in
// the app's folder of lists, or "Untitled list 2.mvd" and so on when that is taken. It is "" when
// the host has no such folder, and the person is asked where to save the list then.
func (m Model) draftPath() string {
	dir := m.host.ListDir
	if dir == "" {
		return ""
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	for n := 1; n < 1000; n++ {
		name := "Untitled list"
		if n > 1 {
			name = fmt.Sprintf("Untitled list %d", n)
		}
		path := filepath.Join(dir, name+playlistfile.Extension)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}

	return ""
}
