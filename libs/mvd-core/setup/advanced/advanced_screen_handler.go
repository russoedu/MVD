package advanced

import (
	"fmt"
	"strconv"
	"strings"
	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/outline"
	"youtube-downloader/libs/mvd-core/settings"

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

const advItemCount = 6

var advLabels = []string{"Extra yt-dlp args", "Parallel fragments per video", "Auto Retry", "Retry cycles", "Retry wait (s x attempt)", "Start timeout (s)"}

// Model is the "if you know what you're doing" screen: raw yt-dlp
// args, the fragment count and auto-retry.
type Model struct {
	cfg           config.Config
	cursor        int
	mode          settings.Mode
	input         textinput.Model
	width, height int
}

func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	return Model{cfg: cfg, input: ti}
}

// Config is the settings as they stand, including changes not yet saved.
func (m Model) Config() config.Config { return m.cfg }

func (m Model) SetSize(w, h int) Model {
	m.width, m.height = w, h
	m.input.SetWidth(max(10, w-8))
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome, config.Config) {
	k, isKey := msg.(tea.KeyPressMsg)

	if m.mode != settings.EditNone {
		if isKey {
			switch k.String() {
			case "esc":
				m.mode = settings.EditNone
				return m, nil, None, m.cfg
			case "enter":
				m.commit()
				m.mode = settings.EditNone
				return m, nil, None, m.cfg
			}
			if m.mode == settings.EditNumber {
				switch k.String() {
				case "up":
					m.input.SetValue(strconv.Itoa(m.numberValue() + 1))
					m.input.CursorEnd()
					return m, nil, None, m.cfg
				case "down":
					if n := m.numberValue(); n > 0 {
						m.input.SetValue(strconv.Itoa(n - 1))
						m.input.CursorEnd()
					}
					return m, nil, None, m.cfg
				}
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
			if m.cursor < advItemCount-1 {
				m.cursor++
			}
		case "s", "ctrl+s":
			return m, nil, Save, m.cfg
		case "esc":
			return m, nil, Cancel, m.cfg
		case "enter", " ":
			m.activate()
		}
	}
	return m, nil, None, m.cfg
}

func (m *Model) activate() {
	switch m.cursor {
	case 0:
		m.input.SetValue(strings.Join(m.cfg.ExtraArgs, " "))
		m.input.CursorEnd()
		m.input.Focus()
		m.mode = settings.EditText
	case 1:
		m.input.SetValue(strconv.Itoa(m.cfg.ConcurrentFragments))
		m.input.CursorEnd()
		m.input.Focus()
		m.mode = settings.EditNumber
	case 2:
		m.cfg.AutoRetry = !m.cfg.AutoRetry
	case 3, 4, 5:
		m.input.SetValue(strconv.Itoa(m.current()))
		m.input.CursorEnd()
		m.input.Focus()
		m.mode = settings.EditNumber
	}
}

func (m Model) numberValue() int {
	n, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || n < 0 {
		return m.current()
	}
	return n
}

// current is the number the setting under the cursor holds now.
func (m Model) current() int {
	switch m.cursor {
	case 3:
		return m.cfg.RetryCycles
	case 4:
		return m.cfg.RetryWaitSeconds
	case 5:
		return m.cfg.StartTimeoutSeconds
	}
	return m.cfg.ConcurrentFragments
}

func (m *Model) commit() {
	switch m.cursor {
	case 0:
		m.cfg.ExtraArgs = strings.Fields(m.input.Value())
	case 1:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n >= 0 {
			m.cfg.ConcurrentFragments = n
		}
	case 3:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n >= 1 {
			m.cfg.RetryCycles = n
		}
	case 4:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n >= 1 {
			m.cfg.RetryWaitSeconds = n
		}
	case 5:
		if n, err := strconv.Atoi(strings.TrimSpace(m.input.Value())); err == nil && n >= 1 {
			m.cfg.StartTimeoutSeconds = n
		}
	}
}

func (m Model) display(i int) string {
	switch i {
	case 0:
		if len(m.cfg.ExtraArgs) == 0 {
			return "(none)"
		}
		return strings.Join(m.cfg.ExtraArgs, " ")
	case 1:
		return strconv.Itoa(m.cfg.ConcurrentFragments)
	case 2:
		return settings.YesNo(m.cfg.AutoRetry)
	case 3:
		return strconv.Itoa(m.cfg.RetryCycles)
	case 4:
		return retryWaitText(m.cfg.RetryWaitSeconds)
	case 5:
		return strconv.Itoa(m.cfg.StartTimeoutSeconds)
	}
	return ""
}

func (m Model) Hints() []keybar.Hint {
	if m.mode != settings.EditNone {
		return []keybar.Hint{{Key: "enter", Desc: "apply"}, {Key: "esc", Desc: "cancel"}}
	}
	return []keybar.Hint{{Key: "↑↓", Desc: "move"}, {Key: "enter", Desc: "edit"}, {Key: "s", Desc: "save"}, {Key: "esc", Desc: "back"}}
}

func (m Model) View(width, height int) string {
	props := Default()
	props.Palette = Palette(look.CurrentPalette())
	props.Hints = keybar.Of[HintsHint](m.Hints())

	var labelW = 0
	for _, l := range advLabels {
		labelW = max(labelW, len(l))
	}
	rows := make([]ItemsRow, 0, len(advLabels))
	for i, label := range advLabels {
		row := ItemsRow{Mark: "  ", Name: label, Value: cells.Truncate(m.display(i), max(10, width-labelW-6))}
		if i == m.cursor {
			row.Mark = "▸ "
			if m.mode == settings.EditNone {
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
	items := make([]outline.Item, len(advLabels))
	for i, label := range advLabels {
		items[i] = outline.Item{Label: label, Value: m.display(i), Selected: i == m.cursor}
	}
	if m.mode != settings.EditNone {
		items[m.cursor].Editing = true
		items[m.cursor].Value = m.input.Value()
	}
	return outline.ScreenOutline{Title: "MVD · Advanced", ItemsLabel: "Settings", Items: items, Keys: outline.KeysOf(m.Hints())}
}

// Select puts the cursor on the setting drawn on row, when none is being edited.
func (m Model) Select(row int) Model {
	if m.mode == settings.EditNone && row >= 0 && row < len(advLabels) {
		m.cursor = row
	}
	return m
}

// retryWaitText is the wait setting with what it comes to: the waits of the first
// retries, since each is the setting times the number of the retry.
func retryWaitText(seconds int) string {
	return fmt.Sprintf("%d  (%ds, %ds, %ds, ...)", seconds, seconds, 2*seconds, 3*seconds)
}
