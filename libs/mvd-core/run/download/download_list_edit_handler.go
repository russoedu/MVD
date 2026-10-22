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
