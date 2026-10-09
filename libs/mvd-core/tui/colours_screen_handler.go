package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

type coloursOutcome int

const (
	colNone coloursOutcome = iota
	colSave
	colCancel
)

var colourLabels = []string{
	"Accent (title)", "Focus (active border)", "Highlight (keys, bars)", "Success",
	"Error", "Dim (borders, hints)", "Text", "Selected row (background)",
}

// coloursModel edits the interface colours. A colour applies as soon as it is
// accepted, so the screen itself previews the choice.
type coloursModel struct {
	cfg           config.Config
	cursor        int
	editing       bool
	status        string
	input         textinput.Model
	width, height int
}

func newColoursModel(cfg config.Config) coloursModel {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 7
	return coloursModel{cfg: cfg, input: ti}
}

func (m coloursModel) setSize(w, h int) coloursModel {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

// slot points at the colour setting of item i.
func (m *coloursModel) slot(i int) *string {
	c := &m.cfg.Colors
	return []*string{&c.Accent, &c.Focus, &c.Highlight, &c.Success, &c.Error, &c.Dim, &c.Text, &c.Selected}[i]
}

func (m coloursModel) defaultFor(i int) string {
	d := config.DefaultThemeColors()
	return []string{d.Accent, d.Focus, d.Highlight, d.Success, d.Error, d.Dim, d.Text, d.Selected}[i]
}

func (m coloursModel) update(msg tea.Msg) (coloursModel, tea.Cmd, coloursOutcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.editing {
		if isKey {
			switch k.String() {
			case "esc":
				m.editing = false
				m.status = ""
				return m, nil, colNone, m.cfg
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if !config.ValidColour(val) {
					m.status = "Use a colour like #ff007f or #f07."
					return m, nil, colNone, m.cfg
				}
				*m.slot(m.cursor) = val
				m.editing = false
				m.status = ""
				ApplyTheme(m.cfg.Colors)
				return m, nil, colNone, m.cfg
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd, colNone, m.cfg
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
			ApplyTheme(m.cfg.Colors)
		case "s", "ctrl+s":
			return m, nil, colSave, m.cfg
		case "esc":
			return m, nil, colCancel, m.cfg
		}
	}
	return m, nil, colNone, m.cfg
}

func (m coloursModel) hints() []keyHint {
	if m.editing {
		return []keyHint{{"enter", "apply"}, {"esc", "cancel"}}
	}
	return []keyHint{{"↑↓", "move"}, {"enter", "edit"}, {"d", "default"}, {"s", "save"}, {"esc", "back"}}
}

func (m coloursModel) display(i int) string {
	return *m.slot(i)
}

func (m coloursModel) view(width, height int) string {
	labelW := 0
	for _, l := range colourLabels {
		labelW = max(labelW, len(l))
	}
	var rows []string
	for i, label := range colourLabels {
		cursor := "  "
		if i == m.cursor {
			cursor = styMagenta.Render("▸ ")
		}
		name := padRight(label, labelW)
		if i == m.cursor && !m.editing {
			name = stySelected.Render(name)
		}
		val := m.display(i)
		shown := styCyan.Render(val)
		if i == m.cursor && m.editing {
			shown = m.input.View()
		}
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("██")
		rows = append(rows, cursor+name+"  "+swatch+" "+shown)
	}
	body := strings.Join(rows, "\n") + "\n\n" +
		styDim.Render("Colours are #rgb or #rrggbb and apply as soon as you accept them.\nSave on the preferences screen to keep them.")
	if m.status != "" {
		body += "\n" + styRed.Render(m.status)
	}

	return screenFrame(width, height, "MVD · Colours", body, keyBar(width, m.hints()))
}
