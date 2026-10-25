package colours

import (
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/setup/coloursession"
)

type Outcome int

const (
	None Outcome = iota
	Save
	Cancel
)

// Model is the colours screen in the terminal: it turns keys into what the session does, owns the
// box the colour is typed in, and draws.
type Model struct {
	session       coloursession.Session
	input         textinput.Model
	width, height int
}

func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 7
	return Model{session: coloursession.New(cfg), input: ti}
}

// Config is the settings as they stand, including changes not yet saved.
func (m Model) Config() config.Config { return m.session.Cfg }

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

// outcome is what the host of the screen is told for what the session asked of it, after the
// colours are drawn with when it asked for that.
func (m Model) outcome(effect coloursession.Effect) Outcome {
	switch effect {
	case coloursession.ApplyColours:
		look.Apply(m.session.Cfg.Colors)
	case coloursession.Save:
		return Save
	case coloursession.Cancel:
		return Cancel
	}
	return None
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.session.Editing {
		if isKey {
			switch k.String() {
			case "esc":
				m.session.CancelEdit()
				return m, nil, None, m.session.Cfg
			case "enter":
				m.session.SetText(m.input.Value())
				effect := m.session.AcceptEdit()
				return m, nil, m.outcome(effect), m.session.Cfg
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
		case "enter", " ":
			m.session.BeginEdit()
			m.input.SetValue(m.session.Text)
			m.input.CursorEnd()
			m.input.Focus()
		case "d":
			m.outcome(m.session.RestoreDefault())
		case "s", "ctrl+s":
			return m, nil, m.outcome(m.session.Keep()), m.session.Cfg
		case "esc":
			return m, nil, m.outcome(m.session.Leave()), m.session.Cfg
		}
	}
	return m, nil, None, m.session.Cfg
}

func (m Model) Hints() []keybar.Hint {
	if m.session.Editing {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "d", Desc: "default"}, {Key: "s", Desc: "save"}, {Key: "esc", Desc: "back"}}
}

func (m Model) View(width, height int) string {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())

	rows := make([]SlotsRow, 0, len(coloursession.Labels))
	for i, label := range coloursession.Labels {
		val := m.session.Value(i)
		row := SlotsRow{Mark: "  ", Name: label, Swatch: val, Value: val}
		if i == m.session.Cursor {
			row.Mark = "▸ "
			if m.session.Editing {
				row.Value = m.input.View()
				row.ValueStyle = CellStyle{Plain: true} // the input draws in the terminal's own colour
			} else {
				row.Style = RowSelected
			}
		}
		rows = append(rows, row)
	}
	props.Slots = rows
	props.Status = m.session.Status
	props.ShowStatus = m.session.Status != ""
	return View(props, width, height).View
}

// Outline describes the screen in plain terms, from the same state and hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	items := make([]outline.Item, len(coloursession.Labels))
	for i, label := range coloursession.Labels {
		items[i] = outline.Item{Label: label, Value: m.session.Value(i), Selected: i == m.session.Cursor}
	}
	if m.session.Editing {
		items[m.session.Cursor].Editing = true
		items[m.session.Cursor].Value = m.session.Text
	}
	return outline.ScreenOutline{Title: "MVD · Colours", ItemsLabel: "Settings", Items: items, Prompt: m.session.Status, Keys: outline.KeysOf(m.Hints())}
}

// Select puts the cursor on the colour drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	m.session.Select(row)
	return m
}
