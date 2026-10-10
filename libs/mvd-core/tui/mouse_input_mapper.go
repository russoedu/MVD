package tui

import (
	"github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
)

// wheelStep is how many rows one notch of the mouse wheel moves.
const wheelStep = 3

func isLeftClick(msg tea.MouseMsg) bool {
	click, ok := msg.(tea.MouseClickMsg)
	return ok && click.Button == tea.MouseLeft
}

// wheelDirection is -1 for a notch up, 1 for a notch down, 0 for anything else.
func wheelDirection(msg tea.MouseMsg) int {
	wheel, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return 0
	}
	switch wheel.Button {
	case tea.MouseWheelUp:
		return -1
	case tea.MouseWheelDown:
		return 1
	}
	return 0
}

// hintAt returns the key bar entry drawn under column x. The bar is a space,
// then each entry as its key, a space and its description, two spaces apart,
// exactly as keyBar and the download screen's key bar draw it.
func hintAt(hints []keyHint, x int) (keyHint, bool) {
	col := 1
	for i, h := range hints {
		if i > 0 {
			col += 2
		}
		width := ansi.StringWidth(h.key) + 1 + ansi.StringWidth(h.desc)
		if x >= col && x < col+width {
			return h, true
		}
		col += width
	}
	return keyHint{}, false
}

// hintKeyMsg is the key press a key bar entry stands for, so a click on the
// entry does what the key does. The arrows that only move are not clickable.
func hintKeyMsg(key string) (tea.KeyPressMsg, bool) {
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

// windowStart is the index of the first row window shows: the same arithmetic,
// so a click can be matched to the row that was drawn there.
func windowStart(n, sel, h int) int {
	if h <= 0 || n <= h {
		return 0
	}
	start := sel - h/2
	if start < 0 {
		start = 0
	}
	if start+h > n {
		start = n - h
	}
	return start
}
