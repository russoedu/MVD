// Package keys makes the key presses the tests of the screens send.
package keys

import tea "charm.land/bubbletea/v2"

// Press is the key press for a name ("enter", "esc", "up", "down", "ctrl+s", ...), or for
// the typed text when the name is not one of those.
func Press(s string) tea.KeyPressMsg {
	switch s {
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	case "ctrl+p":
		return tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}
	case "ctrl+o":
		return tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl}
	case "ctrl+r":
		return tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	case "ctrl+q":
		return tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	default:
		return Text(s)
	}
}

// Text is the key press that types text (its first rune is the key).
func Text(text string) tea.KeyPressMsg {
	r := []rune(text)
	return tea.KeyPressMsg{Code: r[0], Text: text}
}
