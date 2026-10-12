package colours

import (
	"youtube-downloader/libs/mvd-core/connect"
	"strings"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

type Outcome int

const (
	None Outcome = iota
	Save
	Cancel
)

var colourLabels = []string{
	"Accent (title)", "Focus (active border)", "Highlight (keys, bars)", "Success",
	"Error", "Dim (borders, hints)", "Text", "Selected row (background)",
}

// Model edits the interface colours. A colour applies as soon as it is
// accepted, so the screen itself previews the choice.
type Model struct {
	cfg           config.Config
	cursor        int
	editing       bool
	status        string
	input         textinput.Model
	width, height int
}

func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 7
	return Model{cfg: cfg, input: ti}
}

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

// slot points at the colour setting of item i.
func (m *Model) slot(i int) *string {
	c := &m.cfg.Colors
	return []*string{&c.Accent, &c.Focus, &c.Highlight, &c.Success, &c.Error, &c.Dim, &c.Text, &c.Selected}[i]
}

func (m Model) defaultFor(i int) string {
	d := config.DefaultThemeColors()
	return []string{d.Accent, d.Focus, d.Highlight, d.Success, d.Error, d.Dim, d.Text, d.Selected}[i]
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.editing {
		if isKey {
			switch k.String() {
			case "esc":
				m.editing = false
				m.status = ""
				return m, nil, None, m.cfg
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if !config.ValidColour(val) {
					m.status = "Use a colour like #ff007f or #f07."
					return m, nil, None, m.cfg
				}
				*m.slot(m.cursor) = val
				m.editing = false
				m.status = ""
				look.Apply(m.cfg.Colors)
				return m, nil, None, m.cfg
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd, None, m.cfg
	}

	if isKey {
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(colourLabels)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.input.SetValue(*m.slot(m.cursor))
			m.input.CursorEnd()
			m.input.Focus()
			m.editing = true
			m.status = ""
		case "d":
			*m.slot(m.cursor) = m.defaultFor(m.cursor)
			look.Apply(m.cfg.Colors)
		case "s", "ctrl+s":
			return m, nil, Save, m.cfg
		case "esc":
			return m, nil, Cancel, m.cfg
		}
	}
	return m, nil, None, m.cfg
}

func (m Model) Hints() []keybar.Hint {
	if m.editing {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "d", Desc: "default"}, {Key: "s", Desc: "save"}, {Key: "esc", Desc: "back"}}
}

func (m Model) display(i int) string {
	return *m.slot(i)
}

func (m Model) View(width, height int) string {
	return View(m.props(width, height), width, height).View
}

// Live is the page this screen draws and its inputs, by the input names of the design,
// for the live connection to the designer.
func (m Model) Live(width, height int) (page string, inputs map[string]any) {
	return "Colours", connect.Plain(Inputs(m.props(width, height)))
}

// props fills the page for a window of width by height cells.
func (m Model) props(width, height int) Props {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())

	rows := make([]SlotsRow, 0, len(colourLabels))
	for i, label := range colourLabels {
		val := m.display(i)
		row := SlotsRow{Mark: "  ", Name: label, Swatch: val, Value: val}
		if i == m.cursor {
			row.Mark = "▸ "
			if m.editing {
				row.Value = m.input.View()
				row.ValueStyle = CellStyle{Plain: true} // the input draws in the terminal's own colour
			} else {
				row.Style = RowSelected
			}
		}
		rows = append(rows, row)
	}
	props.Slots = rows
	props.Status = m.status
	props.ShowStatus = m.status != ""
	return props
}

// Outline describes the screen in plain terms, from the same state and hints the view draws.
func (m Model) Outline() outline.ScreenOutline {
	items := make([]outline.Item, len(colourLabels))
	for i, label := range colourLabels {
		items[i] = outline.Item{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.editing {
		items[m.cursor].Editing = true
		items[m.cursor].Value = m.input.Value()
	}
	return outline.ScreenOutline{Title: "MVD · Colours", ItemsLabel: "Settings", Items: items, Prompt: m.status, Keys: outline.KeysOf(m.Hints())}
}

// Select puts the cursor on the colour drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	if !m.editing && row >= 0 && row < len(colourLabels) {
		m.cursor = row
	}
	return m
}
