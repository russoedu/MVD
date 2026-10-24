package download

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/pointer"
)

// minColumnWidth is the narrowest the person can make a column.
const minColumnWidth = 4

// columnDrag is the border between two columns of the table being carried: the columns on its two sides
// and how wide each was, and where the pointer was. The border follows the pointer: what one column
// gains the other loses, and the song, which takes what is left, takes what the first of them gives.
type columnDrag struct {
	on          bool
	left, right string
	x           int
	leftW       int
	rightW      int
}

// columnsMenu is the list of columns the person can show or hide, when it is open.
type columnsMenu struct {
	open bool
	sel  int
	// drag is a column being carried by its grip: it is row from, and would go to row to.
	drag     bool
	from, to int
}

// configurable are the columns the person can hide, in the order the menu lists them.
var configurable = []struct{ name, label string }{
	{columnArtist, "Artist"}, {columnVideo, "Video name"}, {columnAlbum, "Album"},
	{columnQuality, "Quality"}, {columnProgress, "Progress"}, {columnStatus, "Status"},
}

// labelOf is how the list of columns calls a column.
func labelOf(name string) string {
	for _, c := range configurable {
		if c.name == name {
			return c.label
		}
	}

	return name
}

// columnOrder is the columns the person can arrange, in the order they are shown: the one they gave,
// else the order of configurable.
func (m Model) columnOrder() []string {
	order := make([]string, 0, len(configurable))
	seen := map[string]bool{}
	for _, name := range m.colOrder {
		if resizable(name) && !seen[name] {
			order, seen[name] = append(order, name), true
		}
	}
	for _, c := range configurable {
		if !seen[c.name] {
			order = append(order, c.name)
		}
	}

	return order
}

// moveColumn puts the column that is at row from of the list at row to.
func (m *Model) moveColumn(from, to int) {
	order := m.columnOrder()
	if from < 0 || from >= len(order) || to < 0 || to >= len(order) || from == to {
		return
	}
	moved := order[from]
	rest := append(append([]string(nil), order[:from]...), order[from+1:]...)
	out := append(append(append([]string(nil), rest[:to]...), moved), rest[to:]...)
	m.colOrder = out
}

// columnSpan is where a column is in the pane of the table, counted from its first cell.
type columnSpan struct {
	name   string
	x0, x1 int
}

// columnSpans are the columns of the table in a pane w cells wide, with where each begins and ends.
func (m Model) columnSpans(w int) []columnSpan {
	var out []columnSpan
	x := 0
	for _, c := range m.tableColumns(w) {
		out = append(out, columnSpan{c.name, x, x + c.width})
		x += c.width + 1
	}

	return out
}

// controlsShown says whether the buttons of the table are there: the playlist has songs, also when the
// filter hides them all.
func (m Model) controlsShown() bool {
	if m.selPlaylist >= len(m.state.Playlists) {
		return false
	}
	pl := m.state.Playlists[m.selPlaylist]

	return pl.Err == "" && pl.Listed && len(pl.Entries) > 0
}

// tableShown says whether the songs are drawn as a table with a row of headings.
func (m Model) tableShown() bool {
	return len(m.visibleEntries()) > 0 && !m.entryNotesShown()
}

// tableInner is the pane of the table's rows: where its first cell is and how wide it is.
func (m Model) tableInner(l downloadLayout) (x, y, w int) {
	inner := l.entries.rows()

	return inner.x, inner.y, inner.w
}

// columnsButton is the button on the top edge of the songs, which opens the columns: where it is.
func (m Model) columnsButton(l downloadLayout) (x, y, w int) {
	const label = "Columns ▾"
	w = len([]rune(label)) + 4
	fx, _, _ := m.filterButton(l)

	return fx - 1 - w, l.entries.y, w
}

// columnsMenuRect is where the list of columns is drawn, under its button.
func (m Model) columnsMenuRect(l downloadLayout) rect {
	w := 24
	h := len(configurable) + 2

	return rect{l.entries.x + l.entries.w - w - 1, l.entries.y + 1, w, h}
}

// withTableControls draws what sits over the table: the buttons for the columns and the filter, and
// the list that is open.
func (m Model) withTableControls(screen string) string {
	if !m.controlsShown() || m.showHelp || m.adding || m.swapping {
		return screen
	}
	l, ok := m.layout()
	if !ok || !l.hasLists {
		return screen
	}
	x, y, w := m.columnsButton(l)
	screen = dialog.Overlay(screen, dialog.WidePill("Columns ▾", w, dialog.PillNormal, false), x, y)
	fx, fy, fw := m.filterButton(l)
	kind := dialog.PillNormal
	if m.filterActive() {
		kind = dialog.PillPrimary
	}
	screen = dialog.Overlay(screen, dialog.WidePill(m.filterLabel(), fw, kind, false), fx, fy)
	screen = m.withFilterMenu(screen, l)
	if !m.columnList.open {
		return screen
	}
	r := m.columnsMenuRect(l)
	order := m.columnOrder()
	lines := make([]string, 0, len(order))
	for i, name := range order {
		mark := "[x]"
		if m.hiddenCols[name] {
			mark = "[ ]"
		}
		grip := "≡"
		if m.columnList.drag && i == m.columnList.to {
			grip = "▶"
		}
		line := grip + " " + mark + " " + labelOf(name)
		if i == m.columnList.sel {
			line = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(look.Colours.Selected)).Render(line + strings.Repeat(" ", max(r.w-2-len([]rune(line)), 0)))
		}
		lines = append(lines, line)
	}

	return dialog.Overlay(screen, dialog.Frame("Columns", lines, r.w, r.h), r.x, r.y)
}

// toggleColumn shows or hides a column.
func (m *Model) toggleColumn(name string) {
	hidden := make(map[string]bool, len(m.hiddenCols)+1)
	for k, v := range m.hiddenCols {
		hidden[k] = v
	}
	hidden[name] = !hidden[name]
	m.hiddenCols = hidden
}

// columnsKey handles a key while the list of columns is open.
func (m Model) columnsKey(key string) Model {
	switch key {
	case "esc", "c", "q":
		m.columnList.open = false
	case "up", "k":
		m.columnList.sel = max(m.columnList.sel-1, 0)
	case "down", "j":
		m.columnList.sel = min(m.columnList.sel+1, len(configurable)-1)
	case "shift+up":
		m.moveColumn(m.columnList.sel, m.columnList.sel-1)
		m.columnList.sel = max(m.columnList.sel-1, 0)
	case "shift+down":
		m.moveColumn(m.columnList.sel, m.columnList.sel+1)
		m.columnList.sel = min(m.columnList.sel+1, len(configurable)-1)
	case "space", "enter":
		m.toggleColumn(m.columnOrder()[m.columnList.sel])
	}

	return m
}

// resizeTo carries the border being dragged to where the pointer is now, so that it follows the pointer
// as far as the columns allow.
func (m Model) resizeTo(x int) Model {
	d := m.colDrag
	_, _, w := m.tableInnerFor()
	dx := x - d.x
	sizes := make(map[string]int, len(m.colWidth)+2)
	for k, v := range m.colWidth {
		sizes[k] = v
	}
	if d.left == columnSong {
		// The columns right of the song are fixed from the right edge: the one on the right of the
		// border grows or shrinks, and the song gives or takes the room.
		song, current := 0, 0
		for _, col := range m.tableColumns(w) {
			switch col.name {
			case columnSong:
				song = col.width
			case d.right:
				current = col.width
			}
		}
		limit := min(current+max(song-minSongWidth, 0), 80)
		sizes[d.right] = min(max(d.rightW-dx, minColumnWidth), limit)
	} else {
		dx = min(max(dx, minColumnWidth-d.leftW), d.rightW-minColumnWidth)
		sizes[d.left], sizes[d.right] = d.leftW+dx, d.rightW-dx
	}
	m.colWidth = sizes

	return m
}

// tableInnerFor is the table's pane, from the layout of the screen as it is.
func (m Model) tableInnerFor() (x, y, w int) {
	l, ok := m.layout()
	if !ok {
		return 0, 0, 0
	}

	return m.tableInner(l)
}

// tableMouse works the mouse on what belongs to the table: the button for the columns and their
// list, the boxes to tick songs, and the borders between the headings, which are carried to resize.
// It says whether the event was for them.
func (m Model) tableMouse(l downloadLayout, msg tea.MouseMsg) (Model, tea.Cmd, bool) {
	if !l.hasLists || !m.controlsShown() {
		return m, nil, false
	}
	pos := msg.Mouse()
	if m.filterList.open {
		return m.filterMouse(l, msg), nil, true
	}

	if m.columnList.open {
		return m.columnsMouse(l, msg), nil, true
	}

	switch msg.(type) {
	case tea.MouseMotionMsg:
		if m.colDrag.on {
			return m.resizeTo(pos.X), nil, true
		}
	case tea.MouseReleaseMsg:
		if m.colDrag.on {
			m.colDrag.on = false

			return m, nil, true
		}
	case tea.MouseClickMsg:
		if !pointer.IsLeftClick(msg) {
			return m, nil, false
		}
		bx, by, bw := m.columnsButton(l)
		if pos.Y == by && pos.X >= bx && pos.X < bx+bw {
			m.columnList, m.filterList = columnsMenu{open: true}, filterMenu{}

			return m, nil, true
		}
		fx, fy, fw := m.filterButton(l)
		if pos.Y == fy && pos.X >= fx && pos.X < fx+fw {
			m.filterList, m.columnList = filterMenu{open: true}, columnsMenu{}

			return m, nil, true
		}
		if !m.tableShown() {
			return m, nil, false
		}
		ix, iy, iw := m.tableInner(l)
		spans := m.columnSpans(iw)
		switch {
		case pos.Y == iy:
			return m.headingClick(spans, ix, pos.X)
		case pos.Y > iy && l.entries.rows().contains(pos.X, pos.Y):
			if _, choosing := m.chooser(); !choosing {
				return m, nil, false
			}
			m.clickEntry(pos.Y)
			en := m.selectedEntry()
			if en == nil {
				return m, nil, true
			}
			for _, s := range spans {
				if s.name == columnCheck && pos.X >= ix+s.x0 && pos.X < ix+s.x1 {
					m.togglePick(en.ID) // the box of a song ticks or unticks that song

					return m, nil, true
				}
			}
			m.pickOnly(en.ID) // anywhere else on the row chooses that song alone

			return m, nil, true
		}
	}

	return m, nil, false
}

// headingClick works a click on the row of headings: the box of the heading ticks every song, and a
// border between two headings is taken to be carried.
func (m Model) headingClick(spans []columnSpan, ix, x int) (Model, tea.Cmd, bool) {
	for i, s := range spans {
		if s.name == columnCheck && x >= ix+s.x0 && x < ix+s.x1 {
			m.togglePickAll()

			return m, nil, true
		}
		if x >= ix+s.x0 && x < ix+s.x1 {
			if sortable(s.name) {
				m.sortOn(s.name) // a click on a heading sorts by it, and again turns the order round
			}

			return m, nil, true
		}
		if x != ix+s.x1 || i+1 >= len(spans) {
			continue
		}
		left, right := s, spans[i+1]
		if !resizable(right.name) || (left.name != columnSong && !resizable(left.name)) {
			return m, nil, true
		}
		m.colDrag = columnDrag{on: true, left: left.name, right: right.name, x: x, leftW: left.x1 - left.x0, rightW: right.x1 - right.x0}

		return m, nil, true
	}

	return m, nil, false
}

// resizable says whether a column can be given another width by the person.
func resizable(name string) bool {
	return strings.Contains(" artist video album quality progress status ", " "+name+" ")
}

// columnsMouse works the mouse on the list of columns: the grip carries a column to another place,
// anywhere else on a row shows or hides it, and a click outside closes the list.
func (m Model) columnsMouse(l downloadLayout, msg tea.MouseMsg) Model {
	r := m.columnsMenuRect(l)
	pos := msg.Mouse()
	row := pos.Y - (r.y + 1)
	inside := r.contains(pos.X, pos.Y) && row >= 0 && row < len(configurable)
	switch msg.(type) {
	case tea.MouseMotionMsg:
		if m.columnList.drag && row >= 0 && row < len(configurable) {
			m.columnList.to = row
		}
	case tea.MouseReleaseMsg:
		if m.columnList.drag {
			m.moveColumn(m.columnList.from, m.columnList.to)
			m.columnList.sel = m.columnList.to
			m.columnList.drag = false
		}
	case tea.MouseClickMsg:
		if !pointer.IsLeftClick(msg) {
			break
		}
		switch {
		case !r.contains(pos.X, pos.Y):
			m.columnList.open = false
		case inside && pos.X <= r.x+3:
			m.columnList = columnsMenu{open: true, sel: row, drag: true, from: row, to: row}
		case inside:
			m.columnList.sel = row
			m.toggleColumn(m.columnOrder()[row])
		}
	}

	return m
}
