package preferences

import (
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/settings"
	folderpage "youtube-downloader/libs/mvd-core/setup/folder"
	"youtube-downloader/libs/mvd-core/setup/preferencessession"

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

const ItemCount = preferencessession.ItemCount

// Model is the preferences screen in the terminal: what it does is the session's. The model turns
// keys into the session's calls, runs what the session asks of the host, owns the box a value is
// typed in and the folder browser, and draws.
type Model struct {
	session       preferencessession.Session
	input         textinput.Model
	folder        folderpage.Model
	pick          FolderPicker // nil when the host has no native chooser
	uninstall     Uninstaller  // nil when the host cannot remove the app
	move          Mover        // nil when the host cannot move the app
	width, height int
}

func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	return Model{session: preferencessession.New(cfg), input: ti}
}

// WithUninstaller offers removing the app on the preferences.
func (m Model) WithUninstaller(uninstall Uninstaller) Model {
	m.uninstall = uninstall
	m.session.CanUninstall = uninstall != nil
	return m
}

// WithMover offers moving the app to its own folder on the preferences.
func (m Model) WithMover(move Mover) Model {
	m.move = move
	m.session.CanMove = move != nil
	return m
}

// WithFolderPicker makes the folder settings open the host's own chooser.
func (m Model) WithFolderPicker(pick FolderPicker) Model {
	m.pick = pick
	m.session.CanPickFolder = pick != nil
	return m
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	m.folder = m.folder.SetSize(w, h)
	return m
}

// run does what the session asked of the host. A question for the person is a command that waits on
// their answer, which comes back as a message of this screen.
func (m *Model) run(effect preferencessession.Effect) (tea.Cmd, Outcome) {
	switch effect.Kind {
	case preferencessession.Save:
		return nil, Save
	case preferencessession.Cancel:
		return nil, Cancel
	case preferencessession.OpenAdvanced:
		return nil, Advanced
	case preferencessession.OpenColours:
		return nil, Colours
	case preferencessession.BrowseFolders:
		m.folder = folderpage.New(effect.Folder).SetSize(m.width, m.height)
	case preferencessession.PickFolder:
		item, pick, start := effect.Item, m.pick, effect.Folder
		return func() tea.Msg {
			path, chosen, err := pick(start)
			return folderPickedMsg{item: item, path: path, chosen: chosen, err: err}
		}, None
	case preferencessession.AskUninstall:
		uninstall := m.uninstall
		return func() tea.Msg {
			outcome, err := uninstall()
			return uninstallAnsweredMsg{outcome: outcome, err: err}
		}, None
	case preferencessession.AskMove:
		move := m.move
		return func() tea.Msg {
			outcome, err := move()
			return moveAnsweredMsg{outcome: outcome, err: err}
		}, None
	}
	return nil, None
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.session.Editing() {
		cmd := m.updateEditor(msg, k, isKey)
		return m, cmd, None, m.session.Cfg
	}

	if isKey {
		m.session.ClearStatus()
		var effect preferencessession.Effect
		switch k.String() {
		case "up", "k":
			m.session.Move(-1)
		case "down", "j":
			m.session.Move(1)
		case "s", "ctrl+s":
			effect = m.session.Keep()
		case "a":
			effect = m.session.OpenAdvanced()
		case "c":
			effect = m.session.OpenColours()
		case "u":
			effect = m.session.AskToUninstall()
		case "m":
			effect = m.session.AskToMove()
		case "esc":
			effect = m.session.Leave()
		case "enter", " ":
			effect = m.session.Activate()
			if m.session.Mode == settings.EditText || m.session.Mode == settings.EditNumber {
				m.beginInput()
			}
		}
		cmd, out := m.run(effect)
		return m, cmd, out, m.session.Cfg
	}
	return m, nil, None, m.session.Cfg
}

// beginInput puts the text the session holds in the box and focuses it.
func (m *Model) beginInput() {
	m.showTyped()
	m.input.Focus()
}

// showTyped puts the text the session holds in the box.
func (m *Model) showTyped() {
	m.input.SetValue(m.session.Text)
	m.input.CursorEnd()
}

func (m *Model) updateEditor(msg tea.Msg, k tea.KeyPressMsg, isKey bool) tea.Cmd {
	switch m.session.Mode {
	case settings.EditConfirming:
		// Nothing to do but wait: the questions are windows of their own.
		switch answered := msg.(type) {
		case uninstallAnsweredMsg:
			m.session.UninstallAnswered(answered.outcome, answered.err)
		case moveAnsweredMsg:
			m.session.MoveAnswered(answered.outcome, answered.err)
		}
		return nil
	case settings.EditPicking:
		// Nothing to do but wait: the chooser is a window of its own.
		if picked, ok := msg.(folderPickedMsg); ok {
			cmd, _ := m.run(m.session.FolderPicked(picked.item, picked.path, picked.chosen, picked.err))
			return cmd
		}
		return nil
	case settings.EditFolder:
		next, cmd, out := m.folder.Update(msg)
		m.folder = next
		switch out {
		case folderpage.Chosen:
			m.session.FolderChosen(m.folder.Dir())
		case folderpage.Cancel:
			m.session.FolderCancelled()
		}
		return cmd
	}
	if isKey {
		switch k.String() {
		case "esc":
			m.session.CancelEdit()
			return nil
		case "enter":
			m.session.SetText(m.input.Value())
			m.session.CommitEdit()
			return nil
		}
		switch m.session.Mode {
		case settings.EditRadio:
			switch k.String() {
			case "up", "k":
				m.session.MoveChoice(-1)
			case "down", "j":
				m.session.MoveChoice(1)
			}
			return nil
		case settings.EditNumber:
			switch k.String() {
			case "up", "down":
				m.session.SetText(m.input.Value())
				m.session.StepNumber(map[string]int{"up": 1, "down": -1}[k.String()])
				m.showTyped()
				return nil
			}
		}
	}
	// This only reassigns the pointer-free copy's input; return a value model.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.session.SetText(m.input.Value())
	return cmd
}

// --- View -------------------------------------------------------------------

func (m Model) View(width, height int) string {
	if m.session.Mode == settings.EditFolder {
		return m.folder.View(width, height)
	}
	labelW := 0
	for _, l := range preferencessession.Labels {
		labelW = max(labelW, len(l))
	}

	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())
	rows := make([]ItemsRow, 0, len(preferencessession.Labels))
	for i, label := range preferencessession.Labels {
		row := ItemsRow{Mark: "  ", Name: label, Value: cells.Truncate(m.session.Display(i), max(10, width-labelW-6))}
		if i == m.session.Cursor {
			row.Mark = "▸ "
			if m.session.Editing() {
				row.Value = cells.Truncate(m.editorView(), max(10, width-labelW-6))
			} else {
				row.Style = RowSelected
			}
		}
		rows = append(rows, row)
	}
	props.Items = rows
	props.ShowTemplateHint = m.session.OnTemplate()
	props.ShowStatus, props.ShowKeys = m.session.Status != "", m.session.Status == ""
	props.Status = " " + m.session.Status
	return View(props, width, height).View
}

func (m Model) Hints() []keybar.Hint {
	mode := m.session.Mode
	if mode == settings.EditFolder {
		return m.folder.Hints()
	}
	if mode == settings.EditPicking || mode == settings.EditConfirming {
		return nil
	}
	if mode != settings.EditNone {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	bar := []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "a", Desc: "advanced"}, {Key: "c", Desc: "colours"}, {Key: "s", Desc: "save"}}
	if m.session.CanMove {
		bar = append(bar, keybar.Hint{Key: "m", Desc: "move app"})
	}
	if m.session.CanUninstall {
		bar = append(bar, keybar.Hint{Key: "u", Desc: "uninstall"})
	}
	return append(bar, keybar.Hint{Key: "esc", Desc: "cancel"})
}

func (m Model) editorView() string {
	switch m.session.Mode {
	case settings.EditRadio:
		var parts []string
		for i, opt := range m.session.RadioOpts {
			if i == m.session.RadioIdx {
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

// Outline describes the screen in plain terms, from the same state and Hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	if m.session.Mode == settings.EditFolder {
		return m.folder.Outline()
	}
	items := make([]outline.Item, len(preferencessession.Labels))
	for i, label := range preferencessession.Labels {
		items[i] = outline.Item{Label: label, Value: m.session.Display(i), Selected: i == m.session.Cursor}
	}
	if m.session.Mode == settings.EditPicking {
		return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: "Choose a folder in the window that opened", Keys: outline.KeysOf(m.Hints())}
	}
	if m.session.Mode == settings.EditConfirming {
		return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.session.Status, Keys: outline.KeysOf(m.Hints())}
	}
	if m.session.Editing() {
		item := &items[m.session.Cursor]
		item.Editing = true
		if m.session.Mode == settings.EditRadio {
			item.Choices, item.Chosen = m.session.RadioOpts, m.session.RadioIdx
			item.Value = m.session.RadioOpts[m.session.RadioIdx]
		} else {
			item.Value = m.session.Text
		}
	}
	return outline.ScreenOutline{Title: "MVD · Preferences", ItemsLabel: "Settings", Items: items, Prompt: m.session.Status, Keys: outline.KeysOf(m.Hints())}
}

// Config is the configuration as edited so far.
func (m Model) Config() config.Config { return m.session.Cfg }

// Mode is how a setting is being changed, if one is.
func (m Model) Mode() settings.Mode { return m.session.Mode }

// Select puts the cursor on the setting drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	m.session.Select(row)
	return m
}

// WithConfig replaces the configuration being edited, as when the advanced or the colours
// screen hands back the one it edited.
func (m Model) WithConfig(cfg config.Config) Model {
	m.session.Cfg = cfg
	return m
}

// Cursor is the index of the setting the cursor is on.
func (m Model) Cursor() int { return m.session.Cursor }
