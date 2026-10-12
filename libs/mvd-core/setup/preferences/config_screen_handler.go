package preferences

import (
	"youtube-downloader/libs/mvd-core/connect"
	"strconv"
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/settings"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

type Outcome int

const (
	None Outcome = iota
	Save
	Cancel
	Advanced
	Colours
)

const ItemCount = 11

var mergeOptions = []string{"mp4", "mkv", "webm"}
var cookieOptions = []string{"all", "off", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari"}

// Model is the preferences screen: a selectable list of settings, each
// edited in place by a toggle, number, text field or radio selector.
type Model struct {
	cfg           config.Config
	cursor        int
	mode          settings.Mode
	input         textinput.Model
	radioOpts     []string
	radioIdx      int
	folder        folderpage.Model
	pick          FolderPicker // nil when the host has no native chooser
	uninstall     Uninstaller  // nil when the host cannot remove the app
	move          Mover        // nil when the host cannot move the app
	status        string       // what happened last, shown in place of the key bar
	width, height int
}

func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	return Model{cfg: cfg, input: ti}
}

// WithUninstaller offers removing the app on the preferences.
func (m Model) WithUninstaller(uninstall Uninstaller) Model {
	m.uninstall = uninstall
	return m
}

// WithMover offers moving the app to its own folder on the preferences.
func (m Model) WithMover(move Mover) Model {
	m.move = move
	return m
}

// WithFolderPicker makes the folder settings open the host's own chooser.
func (m Model) WithFolderPicker(pick FolderPicker) Model {
	m.pick = pick
	return m
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	m.folder = m.folder.SetSize(w, h)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.mode != settings.EditNone {
		cmd := m.updateEditor(msg, k, isKey)
		return m, cmd, None, m.cfg
	}

	if isKey {
		m.status = ""
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < ItemCount-1 {
				m.cursor++
			}
		case "s", "ctrl+s":
			return m, nil, Save, m.cfg
		case "a":
			return m, nil, Advanced, m.cfg
		case "c":
			return m, nil, Colours, m.cfg
		case "u":
			return m, m.askToUninstall(), None, m.cfg
		case "m":
			return m, m.askToMove(), None, m.cfg
		case "esc":
			return m, nil, Cancel, m.cfg
		case "enter", " ":
			return m, m.activate(), None, m.cfg
		}
	}
	return m, nil, None, m.cfg
}

// activate opens the editor for the selected item (or toggles a boolean). A folder
// setting opens the host's chooser when there is one, which the returned command waits on.
func (m *Model) activate() tea.Cmd {
	switch m.cursor {
	case 10:
		m.cfg.SaveNotFound = !m.cfg.SaveNotFound
	case 8:
		m.cfg.CreateLogFile = !m.cfg.CreateLogFile
	case 1, 2, 3, 6, 7:
		m.radioOpts, m.radioIdx = m.radioState()
		m.mode = settings.EditRadio
	case 5:
		m.beginInput(strconv.Itoa(m.cfg.MaxConcurrentDownloads))
		m.mode = settings.EditNumber
	case 0, 9:
		return m.chooseFolder()
	default: // 4 text
		m.beginInput(m.textValue())
		m.mode = settings.EditText
	}
	return nil
}

// askToUninstall starts the questions about removing the app, which the returned
// command waits on; nothing when the host cannot remove the app.
func (m *Model) askToUninstall() tea.Cmd {
	if m.uninstall == nil {
		return nil
	}
	m.mode = settings.EditConfirming
	m.status = "Answer the questions in the window that opened..."
	uninstall := m.uninstall
	return func() tea.Msg {
		outcome, err := uninstall()
		return uninstallAnsweredMsg{outcome: outcome, err: err}
	}
}

// askToMove starts the questions about moving the app to its own folder, which the
// returned command waits on; nothing when the host cannot move the app.
func (m *Model) askToMove() tea.Cmd {
	if m.move == nil {
		return nil
	}
	m.mode = settings.EditConfirming
	m.status = "Answer the questions in the window that opened..."
	move := m.move
	return func() tea.Msg {
		outcome, err := move()
		return moveAnsweredMsg{outcome: outcome, err: err}
	}
}

// finishMoving takes the answer of the move: what to tell the person.
func (m *Model) finishMoving(answered moveAnsweredMsg) {
	m.mode = settings.EditNone
	switch {
	case answered.err != nil:
		m.status = "MVD could not be moved: " + answered.err.Error()
	case answered.outcome == MoveStarted:
		m.status = "MVD was moved to its own folder. This window will close and the new copy will open."
	case answered.outcome == MoveAlreadyThere:
		m.status = "MVD already lives in its own folder."
	case answered.outcome == MoveUnavailable:
		m.status = "This machine has no place to move MVD to, or no way to ask. Start MVD with -move from a terminal."
	default:
		m.status = "MVD was not moved."
	}
}

// finishConfirming takes the answer: what to tell the person.
func (m *Model) finishConfirming(answered uninstallAnsweredMsg) {
	m.mode = settings.EditNone
	switch {
	case answered.err != nil:
		m.status = "The removal could not be asked: " + answered.err.Error()
	case answered.outcome == RemovalStarted:
		m.status = "Removing MVD. This window will close."
	case answered.outcome == RemovalDeclined:
		m.status = "Nothing was removed."
	default:
		m.status = "This machine cannot ask for confirmation, so nothing was removed. Start MVD with -uninstall from a terminal."
	}
}

// folderOfItem is the folder the selected setting holds now.
func (m Model) folderOfItem() string {
	if m.cursor == 0 {
		return m.cfg.OutputDir
	}
	return m.cfg.LogDir
}

// chooseFolder opens the host's chooser, or the built-in folder browser without one.
func (m *Model) chooseFolder() tea.Cmd {
	start := m.folderOfItem()
	if m.pick == nil {
		m.browseFolders(start)
		return nil
	}
	m.mode = settings.EditPicking
	item, pick := m.cursor, m.pick
	return func() tea.Msg {
		path, chosen, err := pick(start)
		return folderPickedMsg{item: item, path: path, chosen: chosen, err: err}
	}
}

func (m *Model) browseFolders(start string) {
	m.folder = folderpage.New(start).SetSize(m.width, m.height)
	m.mode = settings.EditFolder
}

// finishPicking takes the chooser's answer. When it could not be shown, the built-in
// browser is opened instead, so the setting can still be changed.
func (m *Model) finishPicking(picked folderPickedMsg) {
	m.mode = settings.EditNone
	switch {
	case picked.err != nil:
		m.browseFolders(m.folderOfItem())
	case picked.chosen:
		if picked.item == 0 {
			m.cfg.OutputDir = picked.path
		} else {
			m.cfg.LogDir = picked.path
		}
	}
}

func (m *Model) beginInput(val string) {
	m.input.SetValue(val)
	m.input.CursorEnd()
	m.input.Focus()
}

func (m *Model) updateEditor(msg tea.Msg, k tea.KeyPressMsg, isKey bool) tea.Cmd {
	if m.mode == settings.EditConfirming {
		// Nothing to do but wait: the questions are windows of their own.
		switch answered := msg.(type) {
		case uninstallAnsweredMsg:
			m.finishConfirming(answered)
		case moveAnsweredMsg:
			m.finishMoving(answered)
		}
		return nil
	}
	if m.mode == settings.EditPicking {
		// Nothing to do but wait: the chooser is a window of its own.
		if picked, ok := msg.(folderPickedMsg); ok {
			m.finishPicking(picked)
		}
		return nil
	}
	if m.mode == settings.EditFolder {
		next, cmd, out := m.folder.Update(msg)
		m.folder = next
		switch out {
		case folderpage.Chosen:
			if m.cursor == 0 {
				m.cfg.OutputDir = m.folder.Dir()
			} else {
				m.cfg.LogDir = m.folder.Dir()
			}
			m.mode = settings.EditNone
		case folderpage.Cancel:
			m.mode = settings.EditNone
		}
		return cmd
	}
	if isKey {
		switch k.String() {
		case "esc":
			m.mode = settings.EditNone
			return nil
		case "enter":
			m.commit()
			m.mode = settings.EditNone
			return nil
		}
		if m.mode == settings.EditRadio {
			switch k.String() {
			case "up", "k":
				if m.radioIdx > 0 {
					m.radioIdx--
				}
			case "down", "j":
				if m.radioIdx < len(m.radioOpts)-1 {
					m.radioIdx++
				}
			}
			return nil
		}
		if m.mode == settings.EditNumber {
			switch k.String() {
			case "up":
				m.input.SetValue(strconv.Itoa(m.numberValue() + 1))
				m.input.CursorEnd()
				return nil
			case "down":
				if n := m.numberValue(); n > 1 {
					m.input.SetValue(strconv.Itoa(n - 1))
					m.input.CursorEnd()
				}
				return nil
			}
		}
	}
	// This only reassigns the pointer-free copy's input; return a value model.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m Model) numberValue() int {
	n, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || n < 1 {
		return m.cfg.MaxConcurrentDownloads
	}
	return n
}

// commit writes the editor's value back into the config.
func (m *Model) commit() {
	switch m.cursor {
	case 0:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.OutputDir = v
		}
	case 4:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.OutputTemplate = v
		}
	case 9:
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.cfg.LogDir = v
		}
	case 5:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n > 0 {
			m.cfg.MaxConcurrentDownloads = n
		}
	case 1:
		m.cfg.VideoQuality = m.radioOpts[m.radioIdx]
	case 2:
		m.cfg.AudioQuality = m.radioOpts[m.radioIdx]
	case 3:
		m.cfg.MergeOutputFormat = m.radioOpts[m.radioIdx]
	case 6:
		m.cfg.OfficialVideo = config.OfficialMode(m.radioOpts[m.radioIdx])
	case 7:
		applyCookieChoice(&m.cfg, m.radioOpts[m.radioIdx])
	}
}

func applyCookieChoice(cfg *config.Config, opt string) {
	switch opt {
	case "all":
		cfg.AutoCookies, cfg.CookiesFromBrowser = true, ""
	case "off":
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, ""
	default:
		cfg.AutoCookies, cfg.CookiesFromBrowser = false, opt
	}
}

func (m Model) textValue() string {
	switch m.cursor {
	case 0:
		return m.cfg.OutputDir
	case 4:
		return m.cfg.OutputTemplate
	case 9:
		return m.cfg.LogDir
	}
	return ""
}

// radioState returns the options and the current index for the selected item.
func (m Model) radioState() ([]string, int) {
	switch m.cursor {
	case 1:
		return config.VideoPresets, indexOf(config.VideoPresets, m.cfg.VideoQuality)
	case 2:
		return config.AudioPresets, indexOf(config.AudioPresets, m.cfg.AudioQuality)
	case 3:
		return mergeOptions, indexOf(mergeOptions, m.cfg.MergeOutputFormat)
	case 6:
		return config.OfficialModes, indexOf(config.OfficialModes, string(m.cfg.OfficialVideo))
	case 7:
		return cookieOptions, indexOf(cookieOptions, currentCookieChoice(m.cfg))
	}
	return nil, 0
}

func currentCookieChoice(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all"
	default:
		return "off"
	}
}

// --- View -------------------------------------------------------------------

var labels = []string{
	"Output Folder", "Video Quality", "Audio Quality", "Merge Format",
	"Output Template", "Max Concurrent Downloads", "Download Official Music Video",
	"Cookies from Browser", "Create Log File", "Log File Location",
	"Save Songs Without Official Video (CSV)",
}

func (m Model) display(i int) string {
	c := m.cfg
	switch i {
	case 0:
		return c.OutputDir
	case 1:
		return c.VideoQuality
	case 2:
		return c.AudioQuality
	case 3:
		return c.MergeOutputFormat
	case 4:
		return c.OutputTemplate
	case 5:
		return strconv.Itoa(c.MaxConcurrentDownloads)
	case 6:
		return string(c.OfficialVideo)
	case 7:
		return cookieLabel(c)
	case 8:
		return settings.YesNo(c.CreateLogFile)
	case 9:
		return c.LogDir
	case 10:
		if c.OfficialVideo != config.OfficialOnly {
			return settings.YesNo(c.SaveNotFound) + " (used when Download Official is only)"
		}
		return settings.YesNo(c.SaveNotFound)
	}
	return ""
}

func (m Model) View(width, height int) string {
	if m.mode == settings.EditFolder {
		return m.folder.View(width, height)
	}
	return View(m.props(width, height), width, height).View
}

// Live is the page this screen draws and its inputs, by the input names of the design,
// for the live connection to the designer.
func (m Model) Live(width, height int) (page string, inputs map[string]any) {
	if m.mode == settings.EditFolder {
		return m.folder.Live(width, height)
	}
	return "Preferences", connect.Plain(Inputs(m.props(width, height)))
}

// props fills the page for a window of width by height cells.
func (m Model) props(width, height int) Props {
	labelW := 0
	for _, l := range labels {
		labelW = max(labelW, len(l))
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	rows := make([]ItemsRow, 0, len(labels))
	for i, label := range labels {
		row := ItemsRow{Mark: "  ", Name: label, Value: cells.Truncate(m.display(i), max(10, width-labelW-6))}
		if i == m.cursor {
			row.Mark = "▸ "
			if m.mode != settings.EditNone {
				row.Value = cells.Truncate(m.editorView(), max(10, width-labelW-6))
			} else {
				row.Style = RowSelected
			}
		}
		rows = append(rows, row)
	}
	props.Items = rows
	props.ShowTemplateHint = m.cursor == 4
	props.ShowStatus, props.ShowKeys = m.status != "", m.status == ""
	props.Status = " " + m.status
	return props
}

func (m Model) Hints() []keybar.Hint {
	if m.mode == settings.EditFolder {
		return m.folder.Hints()
	}
	if m.mode == settings.EditPicking || m.mode == settings.EditConfirming {
		return nil
	}
	if m.mode != settings.EditNone {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	bar := []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "a", Desc: "advanced"}, {Key: "c", Desc: "colours"}, {Key: "s", Desc: "save"}}
	if m.move != nil {
		bar = append(bar, keybar.Hint{Key: "m", Desc: "move app"})
	}
	if m.uninstall != nil {
		bar = append(bar, keybar.Hint{Key: "u", Desc: "uninstall"})
	}
	return append(bar, keybar.Hint{Key: "esc", Desc: "cancel"})
}

func (m Model) editorView() string {
	switch m.mode {
	case settings.EditRadio:
		var parts []string
		for i, opt := range m.radioOpts {
			if i == m.radioIdx {
				parts = append(parts, look.Yellow.Render("("+opt+")"))
			} else {
				parts = append(parts, look.Dim.Render(opt))
			}
		}
		return strings.Join(parts, " ")
	case settings.EditPicking:
		return look.Dim.Render("choose a folder in the window that opened...")
	default:
		return m.input.View()
	}
}

func cookieLabel(cfg config.Config) string {
	switch {
	case cfg.CookiesFromBrowser != "":
		return cfg.CookiesFromBrowser
	case cfg.AutoCookies:
		return "all (try every browser)"
	default:
		return "off"
	}
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}
	return 0
}

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	if m.mode == settings.EditFolder {
		return m.folder.Outline()
	}
	items := make([]outline.Item, len(labels))
	for i, label := range labels {
		items[i] = outline.Item{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.mode == settings.EditPicking {
		return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: "Choose a folder in the window that opened", Keys: outline.KeysOf(m.Hints())}
	}
	if m.mode == settings.EditConfirming {
		return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.status, Keys: outline.KeysOf(m.Hints())}
	}
	if m.mode != settings.EditNone {
		item := &items[m.cursor]
		item.Editing = true
		if m.mode == settings.EditRadio {
			item.Choices, item.Chosen = m.radioOpts, m.radioIdx
			item.Value = m.radioOpts[m.radioIdx]
		} else {
			item.Value = m.input.Value()
		}
	}
	return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.status, Keys: outline.KeysOf(m.Hints())}
}

// Config is the configuration as edited so far.
func (m Model) Config() config.Config { return m.cfg }

// Mode is how a setting is being changed, if one is.
func (m Model) Mode() settings.Mode { return m.mode }

// Select puts the cursor on the setting drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	if m.mode == settings.EditNone && row >= 0 && row < len(labels) {
		m.cursor = row
	}
	return m
}

// WithConfig replaces the configuration being edited, as when the advanced or the colours
// screen hands back the one it edited.
func (m Model) WithConfig(cfg config.Config) Model {
	m.cfg = cfg
	return m
}

// Cursor is the index of the setting the cursor is on.
func (m Model) Cursor() int { return m.cursor }
