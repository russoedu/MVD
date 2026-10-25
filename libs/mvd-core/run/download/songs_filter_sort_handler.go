package download

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
)

// filterMenu is the list of statuses to show, when it is open.
type filterMenu struct {
	open bool
	sel  int
}

// filterActive says whether some of the statuses are hidden.
func (m Model) filterActive() bool { return m.session.Filter.Active() }

// matchesFilter says whether a song passes the filter by status.
func (m Model) matchesFilter(en *runstate.Entry) bool { return m.session.Filter.Matches(en) }

// toggleStatus shows or hides the songs of a group of statuses.
func (m *Model) toggleStatus(group string) {
	m.session.Filter = m.session.Filter.Toggled(group)
	m.session.SelEntry, m.entriesTop = 0, 0
}

// showOnly shows the songs of one group of statuses and hides the rest; when that is what the filter
// is already, it shows them all.
func (m *Model) showOnly(group string) {
	m.session.Filter = m.session.Filter.Only(group)
	m.session.SelEntry, m.entriesTop = 0, 0
}

// sorted puts the songs in the order the person chose by a click on a heading, else as they are.
func (m *Model) sorted(ids []int) []int { return m.session.Order.Sorted(ids, m.session.State.Entries) }

// sortable says whether a click on the heading of a column sorts the songs by it.
func sortable(name string) bool { return name != columnMark && name != columnCheck }

// sortBy sorts the songs by a column; the column that sorts already changes its direction instead. The
// song the cursor is on stays the one it is on.
func (m *Model) sortOn(column string) {
	keep := -1
	if en := m.selectedEntry(); en != nil {
		keep = en.ID
	}
	m.session.Order = m.session.Order.Clicked(column)
	for i, id := range m.visibleEntries() {
		if id == keep {
			m.session.SelEntry = i
		}
	}
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
	if m.filterActive() {
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
		m.toggleStatus(songlist.Groups[m.filterList.sel].Key)
	case "backspace", "delete":
		m.session.Filter = songlist.Filter{}
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
		m.toggleStatus(songlist.Groups[row].Key)
	}

	return m
}
