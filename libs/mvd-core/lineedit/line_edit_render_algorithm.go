package lineedit

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	plain    = lipgloss.NewStyle().Background(lipgloss.Color("#3b0f2a")).Foreground(lipgloss.Color("#ffffff"))
	selected = lipgloss.NewStyle().Background(lipgloss.Color("#00f0ff")).Foreground(lipgloss.Color("#000000"))
	caretOn  = lipgloss.NewStyle().Background(lipgloss.Color("#ffe600")).Foreground(lipgloss.Color("#000000"))
)

// View draws the text in exactly the width the editor has: the characters from the first that
// shows, the selection lit, the cursor a block, and the rest of the line filled.
func (m Model) View() string {
	lo, hi := m.bounds()
	var b strings.Builder
	used := 0
	for i := m.scroll; i <= len(m.text) && used < m.width; i++ {
		r := ' '
		if i < len(m.text) {
			r = m.text[i]
		} else if i != m.caret {
			break
		}
		cell := string(r)
		w := max(ansi.StringWidth(cell), 1)
		if used+w > m.width {
			break
		}
		switch {
		case i == m.caret && lo == hi:
			b.WriteString(caretOn.Render(cell))
		case i >= lo && i < hi:
			b.WriteString(selected.Render(cell))
		default:
			b.WriteString(plain.Render(cell))
		}
		used += w
	}
	b.WriteString(plain.Render(strings.Repeat(" ", max(m.width-used, 0))))
	return b.String()
}
