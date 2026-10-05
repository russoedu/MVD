package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	m.input.Width = max(10, w-8)
	return m
}

func (m advancedModel) update(msg tea.Msg) (advancedModel, tea.Cmd, advancedOutcome, config.Config) {
	k, isKey := msg.(tea.KeyMsg)

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

func (m advancedModel) hints() []keyHint {
	if m.mode != editNone {
		return []keyHint{{"enter", "apply"}, {"esc", "cancel"}}
	}
	return []keyHint{{"↑↓", "move"}, {"enter", "edit"}, {"s", "save"}, {"esc", "back"}}
}

func (m advancedModel) view(width, height int) string {
	labelW := 0
	for _, l := range advLabels {
		if len(l) > labelW {
			labelW = len(l)
		}
	}
	var rows []string
	for i, label := range advLabels {
		cursor := "  "
		if i == m.cursor {
			cursor = styMagenta.Render("▸ ")
		}
		name := padRight(label, labelW)
		if i == m.cursor && m.mode == editNone {
			name = stySelected.Render(name)
		}
		val := m.display(i)
		if i == m.cursor && m.mode != editNone {
			val = m.input.View()
		}
		rows = append(rows, cursor+name+"  "+styCyan.Render(truncate(val, max(10, width-labelW-6))))
	}
	body := strings.Join(rows, "\n") + "\n\n" +
		styDim.Render("These are passed straight to yt-dlp. Parallel fragments and the cookie\nflags are already handled; only add things you understand.")

	return screenFrame(width, height, "MVD · Advanced", body, keyBar(width, m.hints()))
}
