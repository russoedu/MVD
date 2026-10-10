package tui

import (
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
	"youtube-downloader/libs/mvd-core/sourcelist"
)

// NewAppModel returns the whole interactive app as one Bubble Tea model: the
// setup screens, then a download run when the user starts, then the setup
// screens again, the way the terminal app loops between RunSetup and
// RunDownload. A host that cannot run one program per screen (over the web, for
// instance) runs this model instead. It quits only when the user quits the
// setup screens. The returned model also has an Outline method (see
// ScreenOutline).
func NewAppModel(in AppInput) tea.Model {
	m := appModel{in: in, cfg: in.Setup.Cfg, urls: in.Setup.URLs}
	m.setup = m.newSetup(in.Setup.OpenConfig)
	return m
}

// runClosedMsg says a run finished winding down after the user left its screen.
type runClosedMsg struct{}

// planClosedMsg says the plan-only run of the editor has wound down.
type planClosedMsg struct{}

type appModel struct {
	in            AppInput
	cfg           config.Config
	urls          []string
	width, height int

	setup       setupModel
	download    model
	run         Run
	downloading bool
	stopping    bool   // the run is winding down; input is ignored until it has
	notice      string // why the user is back on the setup screens

	// The playlist editor: its screen, the plan-only run that feeds it while it works, and
	// the songs of the download it started, so the editor can mark them when it ends.
	editor     editorModel
	editing    bool
	planRun    Run
	leaving    bool // the editor was left while its plan run was still winding down
	fromEditor bool
	editorKeys []string
}

func (m appModel) newSetup(openConfig bool) setupModel {
	in := m.in.Setup
	in.Cfg, in.URLs, in.OpenConfig = m.cfg, m.urls, openConfig
	s := newSetupModel(in)
	s.embedded = true
	if m.width == 0 {
		// The size is not known yet. Sizing the screens to nothing would wrap
		// the list into a sliver and leave it scrolled when the real size comes.
		return s
	}
	next, _ := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	return next.(setupModel)
}

func (m appModel) Init() tea.Cmd {
	if m.in.ReviewOnStart && len(m.urls) > 0 {
		res := SetupResult{Action: ActionReview, Cfg: m.cfg, URLs: m.urls}
		return tea.Batch(m.setup.Init(), func() tea.Msg { return setupFinishedMsg{result: res} })
	}
	return m.setup.Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m.forward(msg)
	case setupFinishedMsg:
		return m.startRun(msg.result)
	case downloadFinishedMsg:
		return m.stopRun()
	case runClosedMsg:
		return m.backToSetup()
	case planClosedMsg:
		return m.planClosed()
	case editorPlannedMsg:
		return m.closePlanRun(false)
	case editorLeaveMsg:
		return m.leaveEditor()
	case editorDownloadMsg:
		return m.startEditorDownload(msg)
	case editorEvMsg, editorEvClosed, editorTickMsg, editorDescribedMsg:
		if !m.editing {
			return m, nil
		}
		return m.forwardToEditor(msg)
	}
	if m.stopping {
		return m, nil
	}
	if !m.downloading {
		switch msg.(type) {
		case tickMsg, evMsg, evClosed: // left over from a run that has ended
			return m, nil
		case tea.KeyPressMsg:
			m.notice = ""
		}
	}
	return m.forward(msg)
}

func (m appModel) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editing && !m.downloading {
		return m.forwardToEditor(msg)
	}
	if m.downloading {
		next, cmd := m.download.Update(msg)
		m.download = next.(model)
		return m, cmd
	}
	next, cmd := m.setup.Update(msg)
	m.setup = next.(setupModel)
	return m, cmd
}

// startRun leaves the setup screens: it quits, stays when there is nothing to
// download, or starts a run and shows its screen.
func (m appModel) startRun(res SetupResult) (tea.Model, tea.Cmd) {
	m.cfg, m.urls = res.Cfg, res.URLs
	if res.Action == ActionQuit {
		return m, tea.Quit
	}
	if len(m.urls) == 0 {
		m.setup = m.newSetup(false)
		return m, m.setup.Init()
	}
	// A saved plan in the list is opened in the editor, whichever key started the run.
	if res.Action == ActionReview || len(planFilesOf(m.urls)) > 0 {
		return m.startReview()
	}
	run, err := m.in.Start(m.cfg, m.urls)
	if err != nil {
		m.setup = m.newSetup(false)
		m.notice = "Cannot start the download: " + err.Error()
		return m, m.setup.Init()
	}
	dm := newModel(run)
	dm.embedded = true
	next, _ := dm.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.run, m.download, m.downloading = run, next.(model), true
	return m, m.download.Init()
}

// stopRun stops the run in the background, so the screen keeps drawing while
// the downloads wind down.
func (m appModel) stopRun() (tea.Model, tea.Cmd) {
	m.stopping = true
	run := m.run
	return m, func() tea.Msg {
		run.Close()
		return runClosedMsg{}
	}
}

// backToSetup returns to the setup screens once the run has closed, and
// empties the saved list when every item has been dealt with.
func (m appModel) backToSetup() (tea.Model, tea.Cmd) {
	if m.fromEditor {
		return m.finishEditorDownload()
	}
	if t := m.download.state.Tally(); t.Total > 0 && t.Queued == 0 && t.Running == 0 {
		_ = sourcelist.Clear(m.in.Setup.ListPath)
		m.urls = nil
	}
	m.run, m.downloading, m.stopping = nil, false, false
	m.setup = m.newSetup(false)
	return m, m.setup.Init()
}

func (m appModel) render() string {
	if m.editing && !m.downloading {
		return m.editor.render()
	}
	if m.downloading {
		return m.download.render()
	}
	view := m.setup.render()
	if m.notice == "" || m.width == 0 {
		return view
	}
	lines := strings.Split(view, "\n")
	lines[len(lines)-1] = fit(" "+styRed.Render(m.notice), m.width)
	return strings.Join(lines, "\n")
}

func (m appModel) View() tea.View { return screenView(m.render()) }

// planFilesOf returns the lines of the list that are saved plans.
func planFilesOf(urls []string) []string {
	var files []string
	for _, u := range urls {
		if strings.EqualFold(filepath.Ext(u), playlistfile.Extension) {
			files = append(files, u)
		}
	}
	return files
}

// startReview opens the playlist editor on the list: the saved plans in it are opened as
// they are, and the playlists are planned by a run that downloads nothing.
func (m appModel) startReview() (tea.Model, tea.Cmd) {
	var brought []playlistfile.Entry
	var playlists []string
	files := map[string]bool{}
	for _, f := range planFilesOf(m.urls) {
		files[f] = true
		file, err := playlistfile.Load(f)
		if err != nil {
			return m.reviewRefused("Cannot open " + f + ": " + err.Error())
		}
		brought = playlisteditor.Merge(brought, file.Entries)
	}
	for _, u := range m.urls {
		if !files[u] {
			playlists = append(playlists, u)
		}
	}

	if len(playlists) == 0 {
		m.editor = newEditorFromRows(m.in.Editor, brought)
		return m.showEditor()
	}
	if m.in.Editor.Plan == nil {
		return m.reviewRefused("Reviewing a playlist before downloading is not available here")
	}

	var saved []playlistfile.Entry
	if len(brought) == 0 && m.in.Editor.SessionFile != "" {
		if file, err := playlistfile.Load(m.in.Editor.SessionFile); err == nil {
			for _, e := range file.Entries {
				if !e.Downloaded {
					saved = file.Entries
					break
				}
			}
		}
	}

	run, err := m.in.Editor.Plan(m.cfg, playlists)
	if err != nil {
		return m.reviewRefused("Cannot look up the playlist: " + err.Error())
	}
	m.planRun = run
	m.editor = newEditorPlanning(m.in.Editor, run, brought, saved)
	return m.showEditor()
}

func (m appModel) reviewRefused(notice string) (tea.Model, tea.Cmd) {
	m.setup = m.newSetup(false)
	m.notice = notice
	return m, m.setup.Init()
}

// showEditor puts the editor on the screen, sized like the rest.
func (m appModel) showEditor() (tea.Model, tea.Cmd) {
	next, _ := m.editor.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.editor, m.editing = next, true
	return m, m.editor.Init()
}

func (m appModel) forwardToEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.editor.Update(msg)
	m.editor = next
	return m, cmd
}

// closePlanRun winds down the plan-only run in the background, once the editor has what
// it needs from it (or the person has left).
func (m appModel) closePlanRun(leaving bool) (tea.Model, tea.Cmd) {
	run := m.planRun
	if run == nil {
		return m, nil
	}
	m.planRun = nil
	if leaving {
		m.leaving, m.stopping = true, true
	}
	return m, func() tea.Msg {
		run.Close()
		return planClosedMsg{}
	}
}

func (m appModel) planClosed() (tea.Model, tea.Cmd) {
	if !m.leaving {
		return m, nil
	}
	m.leaving, m.stopping = false, false
	return m.backToSetupFromEditor()
}

// leaveEditor returns to the setup screens, after the plan run is wound down.
func (m appModel) leaveEditor() (tea.Model, tea.Cmd) {
	if m.planRun != nil {
		return m.closePlanRun(true)
	}
	return m.backToSetupFromEditor()
}

func (m appModel) backToSetupFromEditor() (tea.Model, tea.Cmd) {
	m.editing = false
	m.setup = m.newSetup(false)
	return m, m.setup.Init()
}

// startEditorDownload downloads what the editor chose, on the download screen.
func (m appModel) startEditorDownload(msg editorDownloadMsg) (tea.Model, tea.Cmd) {
	if m.in.Editor.PlanDownload == nil {
		return m, nil
	}
	if err := os.MkdirAll(m.cfg.OutputDir, 0o755); err != nil {
		m.editor.flash("Cannot create the download folder: " + err.Error())
		return m, nil
	}
	run, err := m.in.Editor.PlanDownload(m.cfg, msg.plan)
	if err != nil {
		m.editor.flash("Cannot start the download: " + err.Error())
		return m, nil
	}
	dm := newModel(run)
	dm.embedded = true
	next, _ := dm.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.run, m.download, m.downloading = run, next.(model), true
	m.fromEditor, m.editorKeys = true, msg.keys
	return m, m.download.Init()
}

// finishEditorDownload returns to the editor after a download it started, marking the
// songs that were downloaded so the next download skips them.
func (m appModel) finishEditorDownload() (tea.Model, tea.Cmd) {
	var done []string
	for i, key := range m.editorKeys {
		if en := m.download.state.Entry(i); en != nil && (en.State == engine.StateDone || en.State == engine.StateDuplicate) {
			done = append(done, key)
		}
	}
	m.editor = m.editor.withDownloaded(done)
	m.run, m.downloading, m.stopping = nil, false, false
	m.fromEditor, m.editorKeys = false, nil
	return m, nil
}
