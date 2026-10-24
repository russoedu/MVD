package download

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/pointer"
)

// minColumnWidth is the narrowest the person can make a column.
const minColumnWidth = 4

// columnDrag is the border between two columns of the table being carried: the column that takes the
// change, how wide it was, where the pointer was, and which way a move makes it grow.
type columnDrag struct {
	on     bool
	column string
	x      int
	width  int
	sign   int
}

// columnsMenu is the list of columns the person can show or hide, when it is open.
type columnsMenu struct {
	open bool
	sel  int
}

// configurable are the columns the person can hide, in the order the menu lists them.
var configurable = []struct{ name, label string }{
	{columnArtist, "Artist"}, {columnVideo, "Video name"}, {columnAlbum, "Album"},
	{columnQuality, "Quality"}, {columnProgress, "Progress"}, {columnStatus, "Status"},
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

	return l.entries.x + l.entries.w - 2 - w, l.entries.y, w
}

// columnsMenuRect is where the list of columns is drawn, under its button.
func (m Model) columnsMenuRect(l downloadLayout) rect {
	w := 24
	h := len(configurable) + 2

	return rect{l.entries.x + l.entries.w - w - 1, l.entries.y + 1, w, h}
}

// withTableControls draws what sits over the table: the button for the columns, and their list when
// it is open.
func (m Model) withTableControls(screen string) string {
	if !m.tableShown() || m.showHelp || m.adding || m.swapping {
		return screen
	}
	l, ok := m.layout()
	if !ok || !l.hasLists {
		return screen
	}
	x, y, w := m.columnsButton(l)
	screen = dialog.Overlay(screen, dialog.WidePill("Columns ▾", w, dialog.PillNormal, false), x, y)
	if !m.columnList.open {
		return screen
	}
	r := m.columnsMenuRect(l)
	lines := make([]string, 0, len(configurable))
	for i, c := range configurable {
		mark := "[x]"
		if m.hiddenCols[c.name] {
			mark = "[ ]"
		}
		cursor := "  "
		if i == m.columnList.sel {
			cursor = "▸ "
		}
		lines = append(lines, cursor+mark+" "+c.label)
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
	case "space", "enter":
		m.toggleColumn(configurable[m.columnList.sel].name)
	}

	return m
}

// resizeTo carries the column being dragged to where the pointer is now. The song takes what is left,
// and keeps what it needs, so the column cannot grow past that.
func (m Model) resizeTo(x int) Model {
	d := m.colDrag
	_, _, w := m.tableInnerFor()
	current, song := 0, 0
	for _, col := range m.tableColumns(w) {
		switch col.name {
		case d.column:
			current = col.width
		case columnSong:
			song = col.width
		}
	}
	limit := min(current+max(song-minSongWidth, 0), 80)
	width := min(max(d.width+d.sign*(x-d.x), minColumnWidth), limit)
	sizes := make(map[string]int, len(m.colWidth)+1)
	for k, v := range m.colWidth {
		sizes[k] = v
	}
	sizes[d.column] = width
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
	if !l.hasLists || !m.tableShown() {
		return m, nil, false
	}
	pos := msg.Mouse()

	if m.columnList.open {
		r := m.columnsMenuRect(l)
		if _, isClick := msg.(tea.MouseClickMsg); isClick && pointer.IsLeftClick(msg) {
			if i := pos.Y - (r.y + 1); r.contains(pos.X, pos.Y) && i >= 0 && i < len(configurable) {
				m.columnList.sel = i
				m.toggleColumn(configurable[i].name)
			} else if !r.contains(pos.X, pos.Y) {
				m.columnList.open = false
			}
		}

		return m, nil, true
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
			m.columnList = columnsMenu{open: true}

			return m, nil, true
		}
		ix, iy, iw := m.tableInner(l)
		spans := m.columnSpans(iw)
		switch {
		case pos.Y == iy:
			return m.headingClick(spans, ix, pos.X)
		case pos.Y > iy && l.entries.rows().contains(pos.X, pos.Y):
			for _, s := range spans {
				if s.name == columnCheck && pos.X >= ix+s.x0 && pos.X < ix+s.x1 {
					m.clickEntry(pos.Y)
					if en := m.selectedEntry(); en != nil {
						m.togglePick(en.ID)
					}

					return m, nil, true
				}
			}
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
		if x != ix+s.x1 || i+1 >= len(spans) {
			continue
		}
		left, right := s, spans[i+1]
		drag := columnDrag{on: true, x: x}
		if left.name == columnSong {
			drag.column, drag.width, drag.sign = right.name, right.x1-right.x0, -1
		} else {
			drag.column, drag.width, drag.sign = left.name, left.x1-left.x0, 1
		}
		if !resizable(drag.column) {
			return m, nil, true
		}
		m.colDrag = drag

		return m, nil, true
	}

	return m, nil, false
}

// resizable says whether a column can be given another width by the person.
func resizable(name string) bool {
	return strings.Contains(" artist video album quality progress status ", " "+name+" ")
}
