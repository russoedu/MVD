package download

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/runstate"
)

// The groups of statuses the songs can be filtered by: what the person sees of a song, not every
// state of the engine.
const (
	groupChoose      = "choose"
	groupWaiting     = "waiting"
	groupDownloading = "downloading"
	groupSaved       = "saved"
	groupHad         = "had"
	groupMissing     = "missing"
	groupRejected    = "rejected"
	groupFailed      = "failed"
)

var statusGroups = []struct{ key, label string }{
	{groupChoose, "To choose"}, {groupWaiting, "Waiting"}, {groupDownloading, "Downloading"}, {groupSaved, "Saved"},
	{groupHad, "Already had"}, {groupMissing, "No official video"}, {groupRejected, "Rejected"}, {groupFailed, "Failed"},
}

// statusGroup is the group of statuses a song is in.
func statusGroup(en *runstate.Entry) string {
	switch en.State {
	case engine.StatePlanned:
		return groupChoose
	case engine.StateQueued:
		return groupWaiting
	case engine.StateResolving, engine.StateDownloading, engine.StateMerging:
		return groupDownloading
	case engine.StateDone:
		return groupSaved
	case engine.StateDuplicate:
		return groupHad
	case engine.StateNotFound:
		return groupMissing
	case engine.StateSkipped:
		return groupRejected
	case engine.StateFailed:
		return groupFailed
	}

	return ""
}

// filterMenu is the list of statuses to show, when it is open.
type filterMenu struct {
	open bool
	sel  int
}

// filterActive says whether some of the statuses are hidden.
func (m Model) filterActive() bool { return len(m.statusHidden) > 0 }

// matchesFilter says whether a song passes the filter by status.
func (m Model) matchesFilter(en *runstate.Entry) bool {
	return !m.statusHidden[statusGroup(en)]
}

// toggleStatus shows or hides the songs of a group of statuses.
func (m *Model) toggleStatus(group string) {
	hidden := make(map[string]bool, len(m.statusHidden)+1)
	for k, v := range m.statusHidden {
		hidden[k] = v
	}
	if hidden[group] {
		delete(hidden, group)
	} else {
		hidden[group] = true
	}
	m.statusHidden = hidden
	m.selEntry, m.entriesTop = 0, 0
}

// showOnly shows the songs of one group of statuses and hides the rest; when that is what the filter
// is already, it shows them all.
func (m *Model) showOnly(group string) {
	hidden := map[string]bool{}
	for _, g := range statusGroups {
		if g.key != group {
			hidden[g.key] = true
		}
	}
	if len(m.statusHidden) == len(hidden) && !m.statusHidden[group] {
		hidden = nil
	}
	m.statusHidden = hidden
	m.selEntry, m.entriesTop = 0, 0
}

// sorted puts the songs in the order the person chose by a click on a heading, else as they are.
func (m *Model) sorted(ids []int) []int {
	if m.sortBy == "" {
		return ids
	}
	out := append([]int(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool {
		c := m.compare(m.sortBy, m.state.Entries[out[i]], m.state.Entries[out[j]])
		if m.sortDesc {
			return c > 0
		}

		return c < 0
	})

	return out
}

// compare orders two songs by a column: negative when a comes first.
func (m *Model) compare(column string, a, b *runstate.Entry) int {
	text := func(x, y string) int { return strings.Compare(strings.ToLower(x), strings.ToLower(y)) }
	number := func(x, y float64) int {
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}

		return 0
	}
	switch column {
	case columnIdx:
		return number(float64(a.Index), float64(b.Index))
	case columnSong:
		return text(songName(a), songName(b))
	case columnArtist:
		return text(m.whoSings(a), m.whoSings(b))
	case columnVideo:
		return text(entryTitle(a), entryTitle(b))
	case columnAlbum:
		return text(a.NamedAlbum, b.NamedAlbum)
	case columnQuality:
		return number(float64(a.Height*1000+a.AudioKbps), float64(b.Height*1000+b.AudioKbps))
	case columnProgress:
		return number(a.Percent, b.Percent)
	case columnStatus:
		x, _ := entryStatus(a)
		y, _ := entryStatus(b)

		return text(x, y)
	}

	return 0
}

// sortable says whether a click on the heading of a column sorts the songs by it.
func sortable(name string) bool { return name != columnMark && name != columnCheck }

// sortBy sorts the songs by a column; the column that sorts already changes its direction instead. The
// song the cursor is on stays the one it is on.
func (m *Model) sortOn(column string) {
	keep := -1
	if en := m.selectedEntry(); en != nil {
		keep = en.ID
	}
	if m.sortBy == column {
		m.sortDesc = !m.sortDesc
	} else {
		m.sortBy, m.sortDesc = column, false
	}
	for i, id := range m.visibleEntries() {
		if id == keep {
			m.selEntry = i
		}
	}
	m.keepSelectionVisible()
}

// sortMark is what a heading shows when it is the one that sorts.
func (m Model) sortMark(column string) string {
	switch {
	case m.sortBy != column:
		return ""
	case m.sortDesc:
		return "▼"
	}

	return "▲"
}

// filterLabel is what the button for the filter says.
func (m Model) filterLabel() string {
	if m.filterActive() {
		return "Filter (" + string(rune('0'+min(len(m.statusHidden), 9))) + " hidden) ▾"
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
	h := len(statusGroups) + 2

	return rect{l.entries.x + l.entries.w - w - 1, l.entries.y + 1, w, h}
}

// withFilterMenu draws the list of statuses when it is open.
func (m Model) withFilterMenu(screen string, l downloadLayout) string {
	if !m.filterList.open {
		return screen
	}
	r := m.filterMenuRect(l)
	lines := make([]string, 0, len(statusGroups))
	for i, g := range statusGroups {
		mark := "[ ]"
		if !m.statusHidden[g.key] {
			mark = "[x]"
		}
		line := "  " + mark + " " + g.label
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
		m.filterList.sel = min(m.filterList.sel+1, len(statusGroups)-1)
	case "space", "enter":
		m.toggleStatus(statusGroups[m.filterList.sel].key)
	case "backspace", "delete":
		m.statusHidden = nil
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
	case row >= 0 && row < len(statusGroups):
		m.filterList.sel = row
		m.toggleStatus(statusGroups[row].key)
	}

	return m
}
