package download

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/pointer"
)

// ListEditMsg asks the owner of the list to change it, because the person did something to a
// playlist of the screen: Op is "edit" (change the address of playlist Index), "remove"
// (take it out) or "move" (put it where playlist To is).
type ListEditMsg struct {
	Op        string
	Index, To int
}

// WithEditable says whether the list can be changed from the screen, which it cannot while
// the downloads go.
func (m Model) WithEditable(editable bool) Model {
	m.editable = editable
	if !editable {
		m.dragRowOn = false
	}
	return m
}

// SelectedPlaylist is the playlist the cursor is on.
func (m Model) SelectedPlaylist() int { return m.selPlaylist }

// SelectPlaylist puts the cursor on playlist i, as far as there is one.
func (m Model) SelectPlaylist(i int) Model {
	m.selPlaylist = max(min(i, len(m.state.Playlists)-1), 0)
	return m
}

// editClick works a click or a drag on the rows of the playlists when the list can be changed:
// the cross removes, the pencil changes the address, the grip takes hold of the row. It says
// whether the event was for the icons.
func (m Model) editClick(l downloadLayout, msg tea.MouseMsg) (Model, tea.Cmd, bool) {
	if !m.editable || !l.hasLists {
		return m, nil, false
	}
	pos := msg.Mouse()
	switch msg.(type) {
	case tea.MouseMotionMsg:
		if m.dragRowOn {
			if i, ok := PlaylistsAt(m.downloadProps(m.width, m.height), m.width, m.height, pos.Y); ok {
				m.dropRow = i
			}
			return m, nil, true
		}
	case tea.MouseReleaseMsg:
		if m.dragRowOn {
			m.dragRowOn = false
			if m.dropRow != m.dragRow {
				from, to := m.dragRow, m.dropRow
				m.selPlaylist = to
				return m, func() tea.Msg { return ListEditMsg{Op: "move", Index: from, To: to} }, true
			}
			return m, nil, true
		}
	case tea.MouseClickMsg:
		if !pointer.IsLeftClick(msg) || !l.playlists.rows().contains(pos.X, pos.Y) {
			return m, nil, false
		}
		i, ok := PlaylistsAt(m.downloadProps(m.width, m.height), m.width, m.height, pos.Y)
		if !ok || i >= len(m.state.Playlists) {
			return m, nil, false
		}
		m.focus, m.follow, m.selPlaylist, m.selEntry, m.outScroll = paneLists, false, i, 0, 0
		right := l.playlists.x + l.playlists.w
		switch pos.X {
		case right - 3:
			return m, func() tea.Msg { return ListEditMsg{Op: "remove", Index: i} }, true
		case right - 5:
			return m, func() tea.Msg { return ListEditMsg{Op: "edit", Index: i} }, true
		case l.playlists.x + 1:
			m.dragRowOn, m.dragRow, m.dropRow = true, i, i
			return m, nil, true
		}
	}
	return m, nil, false
}

// PlaylistRow is where playlist i is drawn, in the cells of the screen: the column its row
// starts at, its row, and how wide it is. ok is false when the playlist is not on the screen
// (the list is longer than the room, and it is scrolled out of view).
func (m Model) PlaylistRow(i int) (x, y, w int, ok bool) {
	p := m.downloadProps(m.width, m.height)
	r, found := Layout(p, m.width, m.height).RegionNamed("Playlists rows")
	if !found || p.PlaylistsRows == nil {
		return 0, 0, 0, false
	}
	rows, first := p.PlaylistsRows(r.W, r.H)
	k := i - first
	if k < 0 || k >= len(rows) || i >= len(m.state.Playlists) {
		return 0, 0, 0, false
	}
	return r.X, r.Y + k, r.W, true
}

// PlaylistAt says what is at the cell (x, y) of the screen in the rows of the playlists:
// inside is true when it is in them at all, and index is the playlist drawn there, or -1 for
// the empty room under the last one.
func (m Model) PlaylistAt(x, y int) (index int, inside bool) {
	p := m.downloadProps(m.width, m.height)
	r, found := Layout(p, m.width, m.height).RegionNamed("Playlists rows")
	if !found || x < r.X || x >= r.X+r.W || y < r.Y || y >= r.Y+r.H {
		return -1, false
	}
	if i, ok := PlaylistsAt(p, m.width, m.height, y); ok && i < len(m.state.Playlists) {
		return i, true
	}
	return -1, true
}
