package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// mouse handles a click or a wheel notch on the setup screens: a click on a key
// bar entry presses that key, a click on a setting selects it, and the wheel
// moves like the arrow keys.
func (m setupModel) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.height == 0 {
		return m, nil
	}
	if dir := wheelDirection(msg); dir != 0 {
		return m.pressArrow(dir)
	}
	if !isLeftClick(msg) {
		return m, nil
	}
	if msg.Y == m.height-1 {
		if h, found := hintAt(m.hints(), msg.X); found {
			if key, clickable := hintKeyMsg(h.key); clickable {
				return m.Update(key)
			}
		}
		return m, nil
	}
	m.selectSetting(msg.Y)
	return m, nil
}

// hints is the key bar of the screen being shown.
func (m setupModel) hints() []keyHint {
	switch m.screen {
	case screenConfig:
		if m.config.mode == editFolder {
			return m.config.folder.hints()
		}
		return m.config.hints()
	case screenAdvanced:
		return m.advanced.hints()
	default:
		return m.list.hints()
	}
}

// pressArrow presses the arrow key wheelStep times, up for a negative direction.
func (m setupModel) pressArrow(dir int) (tea.Model, tea.Cmd) {
	key := tea.KeyMsg{Type: tea.KeyDown}
	if dir < 0 {
		key = tea.KeyMsg{Type: tea.KeyUp}
	}
	var model tea.Model = m
	var cmds []tea.Cmd
	for i := 0; i < wheelStep; i++ {
		var cmd tea.Cmd
		model, cmd = model.Update(key)
		cmds = append(cmds, cmd)
	}
	return model, tea.Batch(cmds...)
}

// selectSetting selects the setting drawn on screen row y, when the screen is a
// list of settings and none is being edited. The title takes row 0.
func (m *setupModel) selectSetting(y int) {
	row := y - 1
	switch m.screen {
	case screenConfig:
		if m.config.mode == editNone && row >= 0 && row < len(cfgLabels) {
			m.config.cursor = row
		}
	case screenAdvanced:
		if m.advanced.mode == editNone && row >= 0 && row < len(advLabels) {
			m.advanced.cursor = row
		}
	}
}
