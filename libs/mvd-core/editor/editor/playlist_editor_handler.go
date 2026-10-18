package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"youtube-downloader/libs/mvd-core/look"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"
	"youtube-downloader/libs/mvd-core/viewport"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/runstate"
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
)

// Mode is what the editor is doing with the keys.
type Mode int

const (
	Browsing Mode = iota
	// AskPrevious asks whether to bring the review that was saved last time.
	AskPrevious
	// AskDownload asks what to do with the songs nobody has looked at.
	AskDownload
	// Input is a one-line question: an address, a file, or a reason.
	Input
)

type editorInputPurpose int

const (
	inputReplace editorInputPurpose = iota
	inputSave
	inputLoad
	inputNote
)

// Model is the playlist editor: the songs of a plan, what was proposed for each, and
// what the person decides about them, before anything is downloaded.
type Model struct {
	host EditorHost
	rows []playlistfile.Entry
	sel  int
	top  int

	width, height int

	// planning is true while a plan-only run is still working; ctrl, events and state
	// are that run's.
	planning bool
	ctrl     downloadpage.Controller
	events   <-chan interface{}
	state    *runstate.State
	spin     int
	planDone bool // the run has finished while a question was waiting

	// saved is the review of last time, offered while planning; previous is what is
	// brought along when planning ends.
	saved    []playlistfile.Entry
	previous []playlistfile.Entry

	mode    Mode
	input   textinput.Model
	purpose editorInputPurpose
	// noteFor is the row a reason is being asked for.
	noteFor string
	// picking is true while the system's open-file window is up.
	picking bool

	status      string
	statusUntil time.Time
}

// NewFromRows opens the editor on rows that are already planned.
func NewFromRows(host EditorHost, rows []playlistfile.Entry) Model {
	return Model{host: host, rows: rows}
}

// NewPlanning opens the editor while a plan-only run works. brought is what the
// person asked to bring from files; saved is the review of last time, which they are
// asked about when there is nothing else to bring.
func NewPlanning(host EditorHost, run downloadpage.Run, brought, saved []playlistfile.Entry) Model {
	m := Model{
		host:     host,
		planning: true,
		ctrl:     run,
		events:   run.Events(),
		state:    runstate.New(run.Sources()),
		previous: brought,
	}
	if len(brought) == 0 && len(saved) > 0 {
		m.saved = saved
		m.mode = AskPrevious
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if !m.planning {
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
		if !m.planning {
			return m, nil
		}
		m.spin = (m.spin + 1) % len(look.SpinnerFrames)
		return m, editorTick()

	case EvMsg:
		if !m.planning {
			return m, nil
		}
		m.state.Apply(msg.Ev)
		if _, idle := msg.Ev.(engine.EvIdle); idle {
			return m.planFinished()
		}
		return m, waitEditorEvent(m.events)

	case EvClosed:
		return m, nil

	case FilePickedMsg:
		m.picking = false
		switch {
		case msg.err != nil && msg.save:
			// The window could not be shown: the name is typed instead.
			return m.ask(inputSave, "Save as: ", m.defaultSavePath())
		case msg.err != nil:
			return m.ask(inputLoad, "Open the file: ", m.host.SaveDir)
		case msg.chosen && msg.save:
			return m.saveFile(msg.path)
		case msg.chosen:
			return m.loadFile(msg.path)
		}
		return m, nil

	case DescribedMsg:
		for i := range m.rows {
			if m.rows[i].Key() == msg.key && m.rows[i].Decision == playlistfile.DecisionReplaced {
				m.rows[i].ChosenTitle = msg.title
			}
		}
		m.autosave()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.mouse(msg)
	}

	if m.mode == Input {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// planFinished takes the plan from the run that made it, once every song is dealt with.
func (m Model) planFinished() (Model, tea.Cmd) {
	// A question about last time's review is waiting: the plan is taken when it is answered.
	if m.mode == AskPrevious {
		m.planDone = true
		return m, nil
	}
	if planner, ok := m.ctrl.(Planner); ok {
		fresh := playlisteditor.FromPlan(planner.Plan())
		if len(m.previous) > 0 {
			fresh = playlisteditor.Merge(fresh, m.previous)
		}
		m.rows = fresh
	}
	m.planning, m.planDone = false, false
	m.previous, m.saved = nil, nil
	m.sel, m.top = 0, 0
	m.autosave()
	if len(m.rows) == 0 {
		m.Flash("no songs were found in the list")
	} else {
		m.Flash(fmt.Sprintf("%d songs planned: check them, then download", len(m.rows)))
	}
	return m, func() tea.Msg { return PlannedMsg{} }
}

// Flash shows s on the bottom line for a moment.
func (m *Model) Flash(s string) {
	m.status = s
	m.statusUntil = time.Now().Add(4 * time.Second)
}

// autosave keeps the review where the next session can find it.
func (m Model) autosave() {
	if m.host.SessionFile == "" || m.planning {
		return
	}
	_ = playlistfile.Save(m.host.SessionFile, playlistfile.FromSongs(m.rows))
}

// withDownloaded marks the songs a download finished, and saves.
// WithDownloaded marks the songs with these keys as downloaded, so the next download
// skips them.
func (m Model) WithDownloaded(keys []string) Model {
	n := playlisteditor.MarkDownloaded(m.rows, keys...)
	m.autosave()
	switch n {
	case 0:
		m.Flash("nothing was downloaded")
	case 1:
		m.Flash("1 song downloaded; it will be skipped next time")
	default:
		m.Flash(fmt.Sprintf("%d songs downloaded; they will be skipped next time", n))
	}
	return m
}

func (m Model) current() (playlistfile.Entry, bool) {
	if m.sel < 0 || m.sel >= len(m.rows) {
		return playlistfile.Entry{}, false
	}
	return m.rows[m.sel], true
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, func() tea.Msg { return LeaveMsg{} }
	}
	if key == "ctrl+l" {
		return m, tea.ClearScreen
	}
	if m.picking {
		return m, nil // the file window is up
	}

	switch m.mode {
	case AskPrevious:
		return m.answerPrevious(key)
	case AskDownload:
		return m.answerDownload(key)
	case Input:
		return m.updateInput(msg)
	}

	if m.planning {
		if key == "q" || key == "esc" {
			return m, func() tea.Msg { return LeaveMsg{} }
		}
		return m, nil
	}

	switch key {
	case "q", "esc":
		return m, func() tea.Msg { return LeaveMsg{} }
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "pgup":
		m.move(-m.listHeight())
	case "pgdown":
		m.move(m.listHeight())
	case "home":
		m.move(-len(m.rows))
	case "end":
		m.move(len(m.rows))
	case "o", "enter":
		return m.open(false)
	case "v":
		return m.open(true)
	case "c", " ":
		return m.confirm()
	case "u":
		return m.decide(playlistfile.DecisionOriginal, "", "")
	case "x":
		return m.decide(playlistfile.DecisionSkipped, "", "")
	case "z", "backspace":
		return m.decide(playlistfile.DecisionNone, "", "")
	case "r":
		return m.ask(inputReplace, "Address of the video to download instead: ", "")
	case "a":
		return m.confirmAll()
	case "d":
		return m.download()
	case "s":
		if m.host.SaveFile != nil {
			return m.pickSaveFile()
		}
		return m.ask(inputSave, "Save as: ", m.defaultSavePath())
	case "l":
		if m.host.PickFile != nil {
			return m.pickFile()
		}
		return m.ask(inputLoad, "Open the file: ", m.host.SaveDir)
	}
	return m, nil
}

func (m *Model) move(delta int) {
	if len(m.rows) == 0 {
		return
	}
	m.sel = viewport.Clamp(m.sel+delta, 0, len(m.rows)-1)
	m.keepSelectionVisible()
}

func (m *Model) keepSelectionVisible() {
	h := m.listHeight()
	if m.sel < m.top {
		m.top = m.sel
	}
	if m.sel >= m.top+h {
		m.top = m.sel - h + 1
	}
	m.top = viewport.Clamp(m.top, 0, max(0, len(m.rows)-h))
}

// open shows a video in the browser: the one that would be downloaded, or with original
// the upload that is in the playlist.
func (m Model) open(original bool) (Model, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	id, _, has := playlisteditor.Target(row)
	if original || !has {
		id = row.Upload.ID
	}
	if id == "" {
		m.Flash("there is no video to open")
		return m, nil
	}
	if m.host.Open == nil {
		m.Flash(playlisteditor.VideoAddress(id))
		return m, nil
	}
	if err := m.host.Open(playlisteditor.VideoAddress(id)); err != nil {
		m.Flash("cannot open the browser: " + err.Error())
		return m, nil
	}
	m.Flash("opened " + playlisteditor.VideoAddress(id))
	return m, nil
}

// confirm accepts what was proposed for the song and moves on to the next one.
func (m Model) confirm() (Model, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	if _, _, has := playlisteditor.Target(playlistfile.Entry{Upload: row.Upload, TargetID: row.TargetID, Kind: row.Kind}); !has {
		m.Flash("nothing was found for this song: replace it, take the original or skip it")
		return m, nil
	}
	return m.decide(playlistfile.DecisionConfirmed, "", "")
}

// confirmAll accepts what was proposed for every song nobody has decided on.
func (m Model) confirmAll() (Model, tea.Cmd) {
	n := 0
	for i := range m.rows {
		row := m.rows[i]
		if row.Decision != playlistfile.DecisionNone || row.Downloaded {
			continue
		}
		if _, _, has := playlisteditor.Target(row); has {
			m.rows[i].Decision = playlistfile.DecisionConfirmed
			n++
		}
	}
	m.autosave()
	m.Flash(fmt.Sprintf("confirmed %d songs", n))
	return m, nil
}

// decide records what the person chose for the selected song. In a build made for testing
// a decision that goes against the proposal asks why.
func (m Model) decide(decision, chosenID, chosenTitle string) (Model, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	if row.Downloaded {
		m.Flash("this song has been downloaded already")
		return m, nil
	}
	if decision == playlistfile.DecisionOriginal && row.Upload.ID == "" {
		m.Flash("this song has no original upload")
		return m, nil
	}
	m.rows[m.sel].Decision = decision
	m.rows[m.sel].ChosenID, m.rows[m.sel].ChosenTitle = chosenID, chosenTitle
	m.autosave()
	// A decision hands the person the next song, as confirming always has; undoing one
	// stays where it is.
	if decision != playlistfile.DecisionNone && m.sel < len(m.rows)-1 {
		m.move(1)
	}

	var cmd tea.Cmd
	if decision == playlistfile.DecisionReplaced && m.host.Describe != nil {
		key, id, describe := row.Key(), chosenID, m.host.Describe
		cmd = func() tea.Msg {
			title, err := describe(id)
			if err != nil {
				return nil
			}
			return DescribedMsg{key: key, title: title}
		}
	}

	against := decision == playlistfile.DecisionReplaced || decision == playlistfile.DecisionOriginal || decision == playlistfile.DecisionSkipped
	if playlisteditor.DebugBuild && against {
		m.noteFor = row.Key()
		asked, askCmd := m.ask(inputNote, "Why? (Enter to leave it blank): ", "")
		return asked, tea.Batch(cmd, askCmd)
	}
	return m, cmd
}

// ask opens the one-line question.
func (m Model) ask(purpose editorInputPurpose, prompt, value string) (Model, tea.Cmd) {
	in := textinput.New()
	in.Prompt = prompt
	in.CharLimit = 0
	in.SetWidth(max(10, m.width-len(prompt)-4))
	in.SetValue(value)
	in.CursorEnd()
	focus := in.Focus()
	m.input, m.purpose, m.mode = in, purpose, Input
	return m, focus
}

func (m Model) updateInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.purpose == inputNote {
			m.logDecision("")
		}
		m.mode = Browsing
		return m, nil
	case "enter":
		return m.submitInput(strings.TrimSpace(m.input.Value()))
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) submitInput(text string) (Model, tea.Cmd) {
	m.mode = Browsing
	switch m.purpose {
	case inputReplace:
		id, ok := playlisteditor.VideoID(text)
		if !ok {
			m.Flash("that is not the address of a YouTube video")
			return m, nil
		}
		return m.decide(playlistfile.DecisionReplaced, id, "")

	case inputNote:
		m.logDecision(text)
		return m, nil

	case inputSave:
		if text == "" {
			return m, nil
		}
		return m.saveFile(text)

	case inputLoad:
		if text == "" {
			return m, nil
		}
		return m.loadFile(text)
	}
	return m, nil
}

// FilePickedMsg is the answer of the system's open-file window.
type FilePickedMsg struct {
	path   string
	chosen bool
	err    error
	// save is true for the answer of the save-file window.
	save bool
}

// pickSaveFile opens the system's save-file window and waits for it in the background.
func (m Model) pickSaveFile() (Model, tea.Cmd) {
	m.picking = true
	pick := m.host.SaveFile
	dir, name := filepath.Split(m.defaultSavePath())
	return m, func() tea.Msg {
		path, chosen, err := pick(filepath.Clean(dir), name)
		return FilePickedMsg{path: path, chosen: chosen, err: err, save: true}
	}
}

// saveFile writes the review to a .mvd file.
func (m Model) saveFile(path string) (Model, tea.Cmd) {
	if filepath.Ext(path) == "" {
		path += playlistfile.Extension
	}
	if err := playlistfile.Save(path, playlistfile.FromSongs(m.rows)); err != nil {
		m.Flash("cannot save: " + err.Error())
		return m, nil
	}
	m.Flash("saved " + path)
	return m, nil
}

// pickFile opens the system's open-file window and waits for it in the background.
func (m Model) pickFile() (Model, tea.Cmd) {
	m.picking = true
	pick, start := m.host.PickFile, m.host.SaveDir
	return m, func() tea.Msg {
		path, chosen, err := pick(start)
		return FilePickedMsg{path: path, chosen: chosen, err: err}
	}
}

// loadFile replaces the review with the songs of a saved .mvd file.
func (m Model) loadFile(path string) (Model, tea.Cmd) {
	file, err := playlistfile.Load(path)
	if err != nil {
		m.Flash("cannot open it: " + err.Error())
		return m, nil
	}
	m.rows = file.Songs()
	m.sel, m.top = 0, 0
	m.autosave()
	m.Flash(fmt.Sprintf("opened %d songs from %s", len(m.rows), filepath.Base(path)))
	return m, nil
}

// logDecision writes the decision made on the row being asked about, with the reason.
func (m Model) logDecision(note string) {
	if !playlisteditor.DebugBuild || m.host.DecisionLog == "" {
		return
	}
	for _, row := range m.rows {
		if row.Key() == m.noteFor {
			_ = playlisteditor.AppendDecision(m.host.DecisionLog, playlisteditor.RecordFor(row, note))
			return
		}
	}
}

// defaultSavePath is the file offered when saving: the first playlist's name, in the
// folder for saved plans.
func (m Model) defaultSavePath() string {
	name := "playlist"
	if len(m.rows) > 0 && m.rows[0].Playlist != "" {
		name = m.rows[0].Playlist
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 32, strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return filepath.Join(m.host.SaveDir, strings.TrimSpace(b.String())+playlistfile.Extension)
}

// download asks for what to do with the songs nobody has looked at, or starts at once.
func (m Model) download() (Model, tea.Cmd) {
	if m.host.PlanDownload == nil {
		m.Flash("downloading is not available here")
		return m, nil
	}
	if playlisteditor.Unreviewed(m.rows) > 0 {
		m.mode = AskDownload
		return m, nil
	}
	return m.startDownload(false)
}

func (m Model) startDownload(includeUnreviewed bool) (Model, tea.Cmd) {
	planned, keys := playlisteditor.Selection(m.rows, includeUnreviewed)
	if len(planned) == 0 {
		m.Flash("nothing to download: confirm, replace or take the original of some songs first")
		return m, nil
	}
	return m, func() tea.Msg { return DownloadMsg{Plan: planned, Keys: keys} }
}

func (m Model) answerDownload(key string) (Model, tea.Cmd) {
	m.mode = Browsing
	switch key {
	case "y", "Y", "enter":
		return m.startDownload(false)
	case "a", "A":
		return m.startDownload(true)
	}
	return m, nil
}

// answerPrevious brings the review of last time along, or leaves it.
func (m Model) answerPrevious(key string) (Model, tea.Cmd) {
	switch key {
	case "y", "Y", "enter":
		m.previous = m.saved
	case "n", "N", "esc":
		m.previous = nil
	default:
		return m, nil
	}
	m.mode, m.saved = Browsing, nil
	if m.planDone {
		return m.planFinished()
	}
	return m, nil
}

// Owns is whether msg is one of the editor's own messages, which an app model forwards to
// it while it is on the screen.
func Owns(msg tea.Msg) bool {
	switch msg.(type) {
	case EvMsg, EvClosed, TickMsg, DescribedMsg, FilePickedMsg:
		return true
	}
	return false
}

// Mode is what the editor is doing with the keys.
func (m Model) Mode() Mode { return m.mode }

// Planning is whether a plan-only run is still working.
func (m Model) Planning() bool { return m.planning }

// Rows are the songs, with what was proposed and decided for each.
func (m Model) Rows() []playlistfile.Entry { return m.rows }

// Selected is the index of the song the cursor is on.
func (m Model) Selected() int { return m.sel }

// WithInput puts text in the one-line question that is open.
func (m Model) WithInput(text string) Model {
	m.input.SetValue(text)
	return m
}
