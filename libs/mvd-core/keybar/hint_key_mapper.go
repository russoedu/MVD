package keybar

import tea "charm.land/bubbletea/v2"

// KeyMsg is the key press a key bar entry stands for, so a click on the
// entry does what the key does. The arrows that only move are not clickable.
func KeyMsg(key string) (tea.KeyPressMsg, bool) {
	switch key {
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}, true
	case "ctrl+p":
		return tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, true
	case "ctrl+o":
		return tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl}, true
	case "ctrl+e":
		return tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl}, true
	case "ctrl+r":
		return tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}, true
	case "ctrl+q":
		return tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}, true
	case "enter", "⏎":
		return tea.KeyPressMsg{Code: tea.KeyEnter}, true
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}, true
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}, true
	case "→":
		return tea.KeyPressMsg{Code: tea.KeyRight}, true
	case "←":
		return tea.KeyPressMsg{Code: tea.KeyLeft}, true
	case "↑↓":
		return tea.KeyPressMsg{}, false
	}
	if runes := []rune(key); len(runes) == 1 {
		return tea.KeyPressMsg{Code: runes[0], Text: key}, true
	}
	return tea.KeyPressMsg{}, false
}
