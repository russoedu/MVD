package download

// scrolled is the rows that show in a pane h rows tall when it is scrolled down by top rows, and the
// index of the first. The rows stay where the person left them: choosing a row does not move them,
// only the wheel and the keys that go past an edge do (see keepSelectionVisible).
func scrolled[T any](rows []T, top, h int) (visible []T, first int) {
	n := len(rows)
	if h <= 0 || n <= h {
		return rows, 0
	}
	first = min(max(top, 0), n-h)

	return rows[first : first+h], first
}

// scrollTo is the top of a pane h rows tall, with n rows, after row sel has to be in view: it is
// where it was when sel shows already, and moves only as far as it must when it does not.
func scrollTo(top, sel, h, n int) int {
	if h <= 0 {
		return 0
	}
	if sel < top {
		top = sel
	}
	if sel >= top+h {
		top = sel - h + 1
	}

	return min(max(top, 0), max(n-h, 0))
}

// keepSelectionVisible scrolls the panes just far enough for the selected rows to show.
func (m *Model) keepSelectionVisible() {
	l, ok := m.layout()
	if !ok || !l.hasLists {
		return
	}
	m.playlistsTop = scrollTo(m.playlistsTop, m.selPlaylist, l.playlists.rows().h, len(m.state.Playlists))
	header := 0
	if m.tableShown() {
		header = 1
	}
	m.entriesTop = scrollTo(m.entriesTop, m.selEntry, l.entries.rows().h-header, len(m.visibleEntries()))
}

// scrollPane scrolls a pane by rows with the wheel, with the selection where it is.
func (m *Model) scrollPane(entries bool, rows int) {
	l, ok := m.layout()
	if !ok || !l.hasLists {
		return
	}
	if entries {
		header := 0
		if m.tableShown() {
			header = 1
		}
		m.entriesTop = scrollTo(m.entriesTop+rows, m.entriesTop+rows, l.entries.rows().h-header, len(m.visibleEntries()))

		return
	}
	m.playlistsTop = scrollTo(m.playlistsTop+rows, m.playlistsTop+rows, l.playlists.rows().h, len(m.state.Playlists))
}
