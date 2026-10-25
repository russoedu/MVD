package download

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/songlist"
)

// filterMenu is the list of statuses to show, when it is open.
type filterMenu struct {
	open bool
	sel  int
}

// sortOn sorts the songs by a column; the column that sorts already changes its direction instead. The
// song the cursor is on stays the one it is on.
func (m *Model) sortOn(column string) {
	m.session.SortBy(column)
	m.keepSelectionVisible()
}

// sortMark is what a heading shows when it is the one that sorts.
func (m Model) sortMark(column string) string {
	switch m.session.Order.Direction(column) {
	case 1:
		return "▲"
	case -1:
		return "▼"
	}

	return ""
}

// filterLabel is what the button for the filter says.
func (m Model) filterLabel() string {
	if m.session.Filter.Active() {
		return "Filter (" + string(rune('0'+min(m.session.Filter.HiddenCount(), 9))) + " hidden) ▾"
	}

	return "Filter ▾"
}

// filterButton is the button on the top edge of the songs, to the right of the one for the columns.
func (m Model) filterButton(l downloadLayout) (x, y, w int) {
	w = len([]rune(m.filterLabel())) + 4

	return l.entries.x + l.entries.w - 2 - w, l.entries.y, w
}

// filterMenuRect is where the list of statuses is drawn, under its button.
func (m Model) filterMenuRect(l downloadLayout) rect {
	w := 28
	h := len(songlist.Groups) + 2

	return rect{l.entries.x + l.entries.w - w - 1, l.entries.y + 1, w, h}
}

// withFilterMenu draws the list of statuses when it is open.
func (m Model) withFilterMenu(screen string, l downloadLayout) string {
	if !m.filterList.open {
		return screen
	}
	r := m.filterMenuRect(l)
	lines := make([]string, 0, len(songlist.Groups))
	for i, g := range songlist.Groups {
		mark := "[ ]"
		if !m.session.Filter.Hidden(g.Key) {
			mark = "[x]"
		}
		line := "  " + mark + " " + g.Label
		if i == m.filterList.sel {
			line = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(look.Colours.Selected)).Render(line + strings.Repeat(" ", max(r.w-2-len([]rune(line)), 0)))
		}
		lines = append(lines, line)
	}

	return dialog.Overlay(screen, dialog.Frame("Show songs that are", lines, r.w, r.h), r.x, r.y)
}

// filterKey handles a key while the list of statuses is open.
func (m Model) filterKey(key string) Model {
	switch key {
	case "esc", "F", "q":
		m.filterList.open = false
	case "up", "k":
		m.filterList.sel = max(m.filterList.sel-1, 0)
	case "down", "j":
		m.filterList.sel = min(m.filterList.sel+1, len(songlist.Groups)-1)
	case "space", "enter":
		m.session.ToggleStatus(songlist.Groups[m.filterList.sel].Key)
		m.entriesTop = 0
	case "backspace", "delete":
		m.session.ClearFilter()
	}

	return m
}

// filterMouse works the mouse on the list of statuses: a row shows or hides that status, and a click
// outside closes the list.
func (m Model) filterMouse(l downloadLayout, msg tea.MouseMsg) Model {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || !pointer.IsLeftClick(msg) {
		return m
	}
	r := m.filterMenuRect(l)
	row := click.Y - (r.y + 1)
	switch {
	case !r.contains(click.X, click.Y):
		m.filterList.open = false
	case row >= 0 && row < len(songlist.Groups):
		m.filterList.sel = row
		m.session.ToggleStatus(songlist.Groups[row].Key)
			m.entriesTop = 0
	}

	return m
}
