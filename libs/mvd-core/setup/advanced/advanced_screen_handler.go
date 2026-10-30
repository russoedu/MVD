package advanced

import (
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/settings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/setup/advancedsession"
)

// Outcome is what the host does next after a key press on the screen.
type Outcome int

const (
	None Outcome = iota
	Save
	Cancel
)

// Model is the "if you know what you're doing" screen in the terminal: it turns keys into what the
// session does, owns the box a value is typed in, and draws.
type Model struct {
	session       advancedsession.Session
	input         textinput.Model
	width, height int
}

// New returns the screen for the settings in cfg.
func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	return Model{session: advancedsession.New(cfg), input: ti}
}

// Config is the settings as they stand, including changes not yet saved.
func (m Model) Config() config.Config { return m.session.Cfg }

// SetSize tells the screen that it has w columns and h rows. It also fits the typing box to the width.
func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

// outcome is what the host of the screen is told for what the session asked of it.
func outcome(effect advancedsession.Effect) Outcome {
	switch effect {
	case advancedsession.Save:
		return Save
	case advancedsession.Cancel:
		return Cancel
	}
	return None
}

// showTyped puts the text the session holds in the box.
func (m *Model) showTyped() {
	m.input.SetValue(m.session.Text)
	m.input.CursorEnd()
}

// Update handles msg, which is usually a key press, and returns the new model and a command.
// It also returns the outcome for the host and the settings as they stand, saved or not.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.session.Editing() {
		if isKey {
			switch k.String() {
			case "esc":
				m.session.CancelEdit()
				return m, nil, None, m.session.Cfg
			case "enter":
				m.session.SetText(m.input.Value())
				m.session.CommitEdit()
				return m, nil, None, m.session.Cfg
			}
			if m.session.Mode == settings.EditNumber {
				switch k.String() {
				case "up", "down":
					m.session.SetText(m.input.Value())
					m.session.StepNumber(map[string]int{"up": 1, "down": -1}[k.String()])
					m.showTyped()
					return m, nil, None, m.session.Cfg
				}
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.session.SetText(m.input.Value())
		return m, cmd, None, m.session.Cfg
	}

	if isKey {
		switch k.String() {
		case "up", "k":
			m.session.Move(-1)
		case "down", "j":
			m.session.Move(1)
		case "s", "ctrl+s":
			return m, nil, outcome(m.session.Keep()), m.session.Cfg
		case "esc":
			return m, nil, outcome(m.session.Leave()), m.session.Cfg
		case "enter", " ":
			m.session.Activate()
			if m.session.Editing() {
				m.showTyped()
				m.input.Focus()
			}
		}
	}
	return m, nil, None, m.session.Cfg
}

// Hints returns the key hints for the current state, editing a value or moving between settings.
func (m Model) Hints() []keybar.Hint {
	if m.session.Editing() {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "s", Desc: "save"}, {Key: "esc", Desc: "back"}}
}

// View draws the screen in width columns and height rows and returns it as text.
func (m Model) View(width, height int) string {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())

	var labelW = 0
	for _, l := range advancedsession.Labels {
		labelW = max(labelW, len(l))
	}
	rows := make([]ItemsRow, 0, len(advancedsession.Labels))
	for i, label := range advancedsession.Labels {
		row := ItemsRow{Mark: "  ", Name: label, Value: cells.Truncate(m.session.Display(i), max(10, width-labelW-6))}
		if i == m.session.Cursor {
			row.Mark = "▸ "
			if !m.session.Editing() {
				row.Style = RowSelected
			} else {
				row.Value = cells.Truncate(m.input.View(), max(10, width-labelW-6))
			}
		}
		rows = append(rows, row)
	}
	props.Items = rows
	return View(props, width, height).View
}

// Outline describes the screen in plain terms, from the same state and hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	items := make([]outline.Item, len(advancedsession.Labels))
	for i, label := range advancedsession.Labels {
		items[i] = outline.Item{Label: label, Value: m.session.Display(i), Selected: i == m.session.Cursor}
	}
	if m.session.Editing() {
		items[m.session.Cursor].Editing = true
		items[m.session.Cursor].Value = m.session.Text
	}
	return outline.ScreenOutline{Title: "MVD · Advanced", ItemsLabel: "Settings", Items: items, Keys: outline.KeysOf(m.Hints())}
}

// Select puts the cursor on the setting drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	m.session.Select(row)
	return m
}
