package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	editorEvMsg    struct{ ev interface{} }
	editorEvClosed struct{}
	editorTickMsg  time.Time

	// editorPlannedMsg says the plan run is over: the app closes it.
	editorPlannedMsg struct{}
	// editorLeaveMsg says the person left the editor.
	editorLeaveMsg struct{}
	// editorDownloadMsg asks the app to download these songs; keys say which rows they are.
	editorDownloadMsg struct {
		plan []engine.PlannedEntry
		keys []string
	}
	// editorDescribedMsg carries the title of a video the person picked.
	editorDescribedMsg struct{ key, title string }
)

type editorMode int

const (
	editorBrowsing editorMode = iota
	// editorAskPrevious asks whether to bring the review that was saved last time.
	editorAskPrevious
	// editorAskDownload asks what to do with the songs nobody has looked at.
	editorAskDownload
	// editorInput is a one-line question: an address, a file, or a reason.
	editorInput
)

type editorInputPurpose int

const (
	inputReplace editorInputPurpose = iota
	inputSave
	inputLoad
	inputNote
)

// editorModel is the playlist editor: the songs of a plan, what was proposed for each, and
// what the person decides about them, before anything is downloaded.
type editorModel struct {
	host EditorHost
	rows []playlistfile.Entry
	sel  int
	top  int

	width, height int

	// planning is true while a plan-only run is still working; ctrl, events and state
	// are that run's.
	planning bool
	ctrl     Controller
	events   <-chan interface{}
	state    *runstate.State
	spin     int
	planDone bool // the run has finished while a question was waiting

	// saved is the review of last time, offered while planning; previous is what is
	// brought along when planning ends.
	saved    []playlistfile.Entry
	previous []playlistfile.Entry

	mode    editorMode
	input   textinput.Model
	purpose editorInputPurpose
	// noteFor is the row a reason is being asked for.
	noteFor string

	status      string
	statusUntil time.Time
}

// newEditorFromRows opens the editor on rows that are already planned.
func newEditorFromRows(host EditorHost, rows []playlistfile.Entry) editorModel {
	return editorModel{host: host, rows: rows}
}

// newEditorPlanning opens the editor while a plan-only run works. brought is what the
// person asked to bring from files; saved is the review of last time, which they are
// asked about when there is nothing else to bring.
func newEditorPlanning(host EditorHost, run Run, brought, saved []playlistfile.Entry) editorModel {
	m := editorModel{
		host:     host,
		planning: true,
		ctrl:     run,
		events:   run.Events(),
		state:    runstate.New(run.Sources()),
		previous: brought,
	}
	if len(brought) == 0 && len(saved) > 0 {
		m.saved = saved
		m.mode = editorAskPrevious
	}
	return m
}

func (m editorModel) Init() tea.Cmd {
	if !m.planning {
		return nil
	}
	return tea.Batch(waitEditorEvent(m.events), editorTick())
}

func waitEditorEvent(ch <-chan interface{}) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return editorEvClosed{}
		}
		return editorEvMsg{ev}
	}
}

func editorTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return editorTickMsg(t) })
}

func (m editorModel) Update(msg tea.Msg) (editorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, m.width-24))
		return m, nil

	case editorTickMsg:
		if !m.planning {
			return m, nil
		}
		m.spin = (m.spin + 1) % len(spinnerFrames)
		return m, editorTick()

	case editorEvMsg:
		if !m.planning {
			return m, nil
		}
		m.state.Apply(msg.ev)
		if _, idle := msg.ev.(engine.EvIdle); idle {
			return m.planFinished()
		}
		return m, waitEditorEvent(m.events)

	case editorEvClosed:
		return m, nil

	case editorDescribedMsg:
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

	if m.mode == editorInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// planFinished takes the plan from the run that made it, once every song is dealt with.
func (m editorModel) planFinished() (editorModel, tea.Cmd) {
	// A question about last time's review is waiting: the plan is taken when it is answered.
	if m.mode == editorAskPrevious {
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
		m.flash("no songs were found in the list")
	} else {
		m.flash(fmt.Sprintf("%d songs planned: check them, then download", len(m.rows)))
	}
	return m, func() tea.Msg { return editorPlannedMsg{} }
}

func (m *editorModel) flash(s string) {
	m.status = s
	m.statusUntil = time.Now().Add(4 * time.Second)
}

// autosave keeps the review where the next session can find it.
func (m editorModel) autosave() {
	if m.host.SessionFile == "" || m.planning {
		return
	}
	_ = playlistfile.Save(m.host.SessionFile, playlistfile.File{Entries: m.rows})
}

// withDownloaded marks the songs a download finished, and saves.
func (m editorModel) withDownloaded(keys []string) editorModel {
	n := playlisteditor.MarkDownloaded(m.rows, keys...)
	m.autosave()
	switch n {
	case 0:
		m.flash("nothing was downloaded")
	case 1:
		m.flash("1 song downloaded; it will be skipped next time")
	default:
		m.flash(fmt.Sprintf("%d songs downloaded; they will be skipped next time", n))
	}
	return m
}

func (m editorModel) current() (playlistfile.Entry, bool) {
	if m.sel < 0 || m.sel >= len(m.rows) {
		return playlistfile.Entry{}, false
	}
	return m.rows[m.sel], true
}

func (m editorModel) handleKey(msg tea.KeyPressMsg) (editorModel, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, func() tea.Msg { return editorLeaveMsg{} }
	}
	if key == "ctrl+l" {
		return m, tea.ClearScreen
	}

	switch m.mode {
	case editorAskPrevious:
		return m.answerPrevious(key)
	case editorAskDownload:
		return m.answerDownload(key)
	case editorInput:
		return m.updateInput(msg)
	}

	if m.planning {
		if key == "q" || key == "esc" {
			return m, func() tea.Msg { return editorLeaveMsg{} }
		}
		return m, nil
	}

	switch key {
	case "q", "esc":
		return m, func() tea.Msg { return editorLeaveMsg{} }
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
		return m.ask(inputSave, "Save as: ", m.defaultSavePath())
	case "l":
		return m.ask(inputLoad, "Open the file: ", m.host.SaveDir)
	}
	return m, nil
}

func (m *editorModel) move(delta int) {
	if len(m.rows) == 0 {
		return
	}
	m.sel = clamp(m.sel+delta, 0, len(m.rows)-1)
	m.keepSelectionVisible()
}

func (m *editorModel) keepSelectionVisible() {
	h := m.listHeight()
	if m.sel < m.top {
		m.top = m.sel
	}
	if m.sel >= m.top+h {
		m.top = m.sel - h + 1
	}
	m.top = clamp(m.top, 0, max(0, len(m.rows)-h))
}

// open shows a video in the browser: the one that would be downloaded, or with original
// the upload that is in the playlist.
func (m editorModel) open(original bool) (editorModel, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	id, _, has := playlisteditor.Target(row)
	if original || !has {
		id = row.Upload.ID
	}
	if id == "" {
		m.flash("there is no video to open")
		return m, nil
	}
	if m.host.Open == nil {
		m.flash(playlisteditor.VideoAddress(id))
		return m, nil
	}
	if err := m.host.Open(playlisteditor.VideoAddress(id)); err != nil {
		m.flash("cannot open the browser: " + err.Error())
		return m, nil
	}
	m.flash("opened " + playlisteditor.VideoAddress(id))
	return m, nil
}

// confirm accepts what was proposed for the song and moves on to the next one.
func (m editorModel) confirm() (editorModel, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	if _, _, has := playlisteditor.Target(playlistfile.Entry{Upload: row.Upload, TargetID: row.TargetID, Kind: row.Kind}); !has {
		m.flash("nothing was found for this song: replace it, take the original or skip it")
		return m, nil
	}
	next, cmd := m.decide(playlistfile.DecisionConfirmed, "", "")
	if next.sel < len(next.rows)-1 {
		next.move(1)
	}
	return next, cmd
}

// confirmAll accepts what was proposed for every song nobody has decided on.
func (m editorModel) confirmAll() (editorModel, tea.Cmd) {
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
	m.flash(fmt.Sprintf("confirmed %d songs", n))
	return m, nil
}

// decide records what the person chose for the selected song. In a build made for testing
// a decision that goes against the proposal asks why.
func (m editorModel) decide(decision, chosenID, chosenTitle string) (editorModel, tea.Cmd) {
	row, ok := m.current()
	if !ok {
		return m, nil
	}
	if row.Downloaded {
		m.flash("this song has been downloaded already")
		return m, nil
	}
	if decision == playlistfile.DecisionOriginal && row.Upload.ID == "" {
		m.flash("this song has no original upload")
		return m, nil
	}
	m.rows[m.sel].Decision = decision
	m.rows[m.sel].ChosenID, m.rows[m.sel].ChosenTitle = chosenID, chosenTitle
	m.autosave()

	var cmd tea.Cmd
	if decision == playlistfile.DecisionReplaced && m.host.Describe != nil {
		key, id, describe := row.Key(), chosenID, m.host.Describe
		cmd = func() tea.Msg {
			title, err := describe(id)
			if err != nil {
				return nil
			}
			return editorDescribedMsg{key: key, title: title}
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
func (m editorModel) ask(purpose editorInputPurpose, prompt, value string) (editorModel, tea.Cmd) {
	in := textinput.New()
	in.Prompt = prompt
	in.CharLimit = 0
	in.SetWidth(max(10, m.width-len(prompt)-4))
	in.SetValue(value)
	in.CursorEnd()
	focus := in.Focus()
	m.input, m.purpose, m.mode = in, purpose, editorInput
	return m, focus
}

func (m editorModel) updateInput(msg tea.KeyPressMsg) (editorModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.purpose == inputNote {
			m.logDecision("")
		}
		m.mode = editorBrowsing
		return m, nil
	case "enter":
		return m.submitInput(strings.TrimSpace(m.input.Value()))
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m editorModel) submitInput(text string) (editorModel, tea.Cmd) {
	m.mode = editorBrowsing
	switch m.purpose {
	case inputReplace:
		id, ok := playlisteditor.VideoID(text)
		if !ok {
			m.flash("that is not the address of a YouTube video")
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
		if filepath.Ext(text) == "" {
			text += playlistfile.Extension
		}
		if err := playlistfile.Save(text, playlistfile.File{Entries: m.rows}); err != nil {
			m.flash("cannot save: " + err.Error())
			return m, nil
		}
		m.flash("saved " + text)
		return m, nil

	case inputLoad:
		if text == "" {
			return m, nil
		}
		file, err := playlistfile.Load(text)
		if err != nil {
			m.flash("cannot open it: " + err.Error())
			return m, nil
		}
		m.rows = file.Entries
		m.sel, m.top = 0, 0
		m.autosave()
		m.flash(fmt.Sprintf("opened %d songs from %s", len(m.rows), filepath.Base(text)))
		return m, nil
	}
	return m, nil
}

// logDecision writes the decision made on the row being asked about, with the reason.
func (m editorModel) logDecision(note string) {
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
func (m editorModel) defaultSavePath() string {
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
func (m editorModel) download() (editorModel, tea.Cmd) {
	if m.host.PlanDownload == nil {
		m.flash("downloading is not available here")
		return m, nil
	}
	if playlisteditor.Unreviewed(m.rows) > 0 {
		m.mode = editorAskDownload
		return m, nil
	}
	return m.startDownload(false)
}

func (m editorModel) startDownload(includeUnreviewed bool) (editorModel, tea.Cmd) {
	planned, keys := playlisteditor.Selection(m.rows, includeUnreviewed)
	if len(planned) == 0 {
		m.flash("nothing to download: confirm, replace or take the original of some songs first")
		return m, nil
	}
	return m, func() tea.Msg { return editorDownloadMsg{plan: planned, keys: keys} }
}

func (m editorModel) answerDownload(key string) (editorModel, tea.Cmd) {
	m.mode = editorBrowsing
	switch key {
	case "y", "Y", "enter":
		return m.startDownload(false)
	case "a", "A":
		return m.startDownload(true)
	}
	return m, nil
}

// answerPrevious brings the review of last time along, or leaves it.
func (m editorModel) answerPrevious(key string) (editorModel, tea.Cmd) {
	switch key {
	case "y", "Y", "enter":
		m.previous = m.saved
	case "n", "N", "esc":
		m.previous = nil
	default:
		return m, nil
	}
	m.mode, m.saved = editorBrowsing, nil
	if m.planDone {
		return m.planFinished()
	}
	return m, nil
}
