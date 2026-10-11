package tui

import (
	tea "charm.land/bubbletea/v2"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/pointer"
)

// mouse handles a click or a wheel notch on the setup screens: a click on a key
// bar entry presses that key, a click on a setting selects it, and the wheel
// moves like the arrow keys.
func (m setupModel) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.height == 0 {
		return m, nil
	}
	if dir := pointer.WheelDirection(msg); dir != 0 {
		return m.pressArrow(dir)
	}
	if !pointer.IsLeftClick(msg) {
		return m, nil
	}
	if msg.Mouse().Y == m.height-1 {
		if h, found := keybar.At(m.hints(), msg.Mouse().X); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				return m.Update(key)
			}
		}
		return m, nil
	}
	m.selectSetting(msg.Mouse().Y)
	return m, nil
}

// hints is the key bar of the screen being shown.
func (m setupModel) hints() []keybar.Hint {
	switch m.screen {
	case screenConfig:
		return m.config.Hints()
	case screenAdvanced:
		return m.advanced.Hints()
	case screenColours:
		return m.colours.Hints()
	case screenSongs:
		return m.songs.Hints()
	default:
		return m.list.Hints()
	}
}

// pressArrow presses the arrow key pointer.WheelStep times, up for a negative direction.
func (m setupModel) pressArrow(dir int) (tea.Model, tea.Cmd) {
	key := tea.KeyPressMsg{Code: tea.KeyDown}
	if dir < 0 {
		key = tea.KeyPressMsg{Code: tea.KeyUp}
	}
	var model tea.Model = m
	var cmds []tea.Cmd
	for i := 0; i < pointer.WheelStep; i++ {
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
		m.config = m.config.Select(row)
	case screenAdvanced:
		m.advanced = m.advanced.Select(row)
	case screenColours:
		m.colours = m.colours.Select(row)
	}
}
