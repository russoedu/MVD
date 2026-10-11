package tui

import (
	"strconv"
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	advancedpage "youtube-downloader/libs/mvd-core/setup/advanced"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

type advancedOutcome int

const (
	advNone advancedOutcome = iota
	advSave
	advCancel
)

const advItemCount = 3

var advLabels = []string{"Extra yt-dlp args", "Parallel fragments per video", "Auto Retry"}

// advancedModel is the "if you know what you're doing" screen: raw yt-dlp
// args, the fragment count and auto-retry.
type advancedModel struct {
	cfg           config.Config
	cursor        int
	mode          int
	input         textinput.Model
	width, height int
}

func newAdvancedModel(cfg config.Config) advancedModel {
	ti := textinput.New()
	ti.Prompt = "> "
	return advancedModel{cfg: cfg, input: ti}
}

func (m advancedModel) setSize(w, h int) advancedModel {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

func (m advancedModel) update(msg tea.Msg) (advancedModel, tea.Cmd, advancedOutcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.mode != editNone {
		if isKey {
			switch k.String() {
			case "esc":
				m.mode = editNone
				return m, nil, advNone, m.cfg
			case "enter":
				m.commit()
				m.mode = editNone
				return m, nil, advNone, m.cfg
			}
			if m.mode == editNumber {
				switch k.String() {
				case "up":
					m.input.SetValue(strconv.Itoa(m.numberValue() + 1))
					m.input.CursorEnd()
					return m, nil, advNone, m.cfg
				case "down":
					if n := m.numberValue(); n > 0 {
						m.input.SetValue(strconv.Itoa(n - 1))
						m.input.CursorEnd()
					}
					return m, nil, advNone, m.cfg
				}
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd, advNone, m.cfg
	}

	if isKey {
		switch k.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < advItemCount-1 {
				m.cursor++
			}
		case "s", "ctrl+s":
			return m, nil, advSave, m.cfg
		case "esc":
			return m, nil, advCancel, m.cfg
		case "enter", " ":
			m.activate()
		}
	}
	return m, nil, advNone, m.cfg
}

func (m *advancedModel) activate() {
	switch m.cursor {
	case 0:
		m.input.SetValue(strings.Join(m.cfg.ExtraArgs, " "))
		m.input.CursorEnd()
		m.input.Focus()
		m.mode = editText
	case 1:
		m.input.SetValue(strconv.Itoa(m.cfg.ConcurrentFragments))
		m.input.CursorEnd()
		m.input.Focus()
		m.mode = editNumber
	case 2:
		m.cfg.AutoRetry = !m.cfg.AutoRetry
	}
}

func (m advancedModel) numberValue() int {
	n, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || n < 0 {
		return m.cfg.ConcurrentFragments
	}
	return n
}

func (m *advancedModel) commit() {
	switch m.cursor {
	case 0:
		m.cfg.ExtraArgs = strings.Fields(m.input.Value())
	case 1:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n >= 0 {
			m.cfg.ConcurrentFragments = n
		}
	}
}

func (m advancedModel) display(i int) string {
	switch i {
	case 0:
		if len(m.cfg.ExtraArgs) == 0 {
			return "(none)"
		}
		return strings.Join(m.cfg.ExtraArgs, " ")
	case 1:
		return strconv.Itoa(m.cfg.ConcurrentFragments)
	case 2:
		return yesNo(m.cfg.AutoRetry)
	}
	return ""
}

func (m advancedModel) hints() []keybar.Hint {
	if m.mode != editNone {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "s", Desc: "save"}, {Key: "esc", Desc: "back"}}
}

func (m advancedModel) view(width, height int) string {
	props := advancedpage.Default()
	props.Palette = advancedpage.Palette(look.CurrentPalette())
	props.Hints = keybar.Of[advancedpage.HintsHint](m.hints())

	var labelW = 0
	for _, l := range advLabels {
		labelW = max(labelW, len(l))
	}
	rows := make([]advancedpage.ItemsRow, 0, len(advLabels))
	for i, label := range advLabels {
		row := advancedpage.ItemsRow{Mark: "  ", Name: label, Value: cells.Truncate(m.display(i), max(10, width-labelW-6))}
		if i == m.cursor {
			row.Mark = "▸ "
			if m.mode == editNone {
				row.Style = advancedpage.RowSelected
			} else {
				row.Value = cells.Truncate(m.input.View(), max(10, width-labelW-6))
			}
		}
		rows = append(rows, row)
	}
	props.Items = rows
	return advancedpage.View(props, width, height).View
}
