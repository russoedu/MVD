package tui

import (
	"github.com/charmbracelet/x/ansi"

	tea "github.com/charmbracelet/bubbletea"
)

// wheelStep is how many rows one notch of the mouse wheel moves.
const wheelStep = 3

func isLeftClick(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
}

// wheelDirection is -1 for a notch up, 1 for a notch down, 0 for anything else.
func wheelDirection(msg tea.MouseMsg) int {
	if msg.Action != tea.MouseActionPress {
		return 0
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return -1
	case tea.MouseButtonWheelDown:
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
func hintKeyMsg(key string) (tea.KeyMsg, bool) {
	switch key {
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}, true
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}, true
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}, true
	case "ctrl+q":
		return tea.KeyMsg{Type: tea.KeyCtrlQ}, true
	case "enter", "⏎":
		return tea.KeyMsg{Type: tea.KeyEnter}, true
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}, true
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}, true
	case "→":
		return tea.KeyMsg{Type: tea.KeyRight}, true
	case "←":
		return tea.KeyMsg{Type: tea.KeyLeft}, true
	case "↑↓":
		return tea.KeyMsg{}, false
	}
	if runes := []rune(key); len(runes) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}, true
	}
	return tea.KeyMsg{}, false
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
