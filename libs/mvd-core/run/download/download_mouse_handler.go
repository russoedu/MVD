package download

import (
	tea "charm.land/bubbletea/v2"
	"youtube-downloader/libs/mvd-core/keybar"
	"youtube-downloader/libs/mvd-core/pointer"
)

// rect is a region of the screen, in cells.
type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }

// rows is the rect's inside, below the box title and above its bottom edge.
func (r rect) rows() rect { return rect{r.x + 1, r.y + 1, r.w - 2, r.h - 2} }

// downloadLayout is where View draws each pane.
type downloadLayout struct {
	playlists, entries, output rect
	hasLists, hasOutput        bool
}

// layout is where the screen's design puts each pane, so a mouse position can be matched
// to the pane that was drawn there. ok is false while the screen shows no panes.
func (m Model) layout() (l downloadLayout, ok bool) {
	if m.width < 60 || m.height < 16 || m.session.ShowHelp {
		return l, false
	}
	p := m.downloadProps(m.width, m.height)
	frame := Layout(p, m.width, m.height)

	box := func(name string) rect {
		r, _ := frame.RegionNamed(name)
		return rect{r.X, r.Y, r.W, r.H}
	}
	l.hasLists, l.hasOutput = p.LeftWidth > 0, m.session.FullLog || m.width >= minWideWidth
	if l.hasLists {
		l.playlists, l.entries = box("Playlists box"), box("Entries box")
	}
	if l.hasOutput {
		l.output = box("Output box")
	}
	return l, true
}

// mouse handles a click or a wheel notch on the download screen.
func (m Model) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.session.ShowHelp {
		if pointer.IsLeftClick(msg) {
			m.session.ShowHelp = false
		}
		return m, nil
	}
	if m.session.ConfirmQuit || m.session.Adding {
		return m, nil
	}
	l, ok := m.layout()
	if !ok {
		return m, nil
	}

	if dir := pointer.WheelDirection(msg); dir != 0 {
		return m.wheel(l, msg, dir)
	}
	if next, handled := m.dragBorder(msg); handled {
		return next, nil
	}
	if next, cmd, handled := m.tableMouse(l, msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.editClick(l, msg); handled {
		return next, cmd
	}
	if !pointer.IsLeftClick(msg) {
		return m, nil
	}
	if msg.Mouse().Y == m.height-1 && !m.session.Editable {
		if h, found := keybar.At(m.hints(), msg.Mouse().X); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				return m.handleKey(key)
			}
		}
		return m, nil
	}
	if next, cmd, handled := m.clickChoice(msg.Mouse().X, msg.Mouse().Y); handled {
		return next, cmd
	}
	if l.hasLists && l.playlists.rows().contains(msg.Mouse().X, msg.Mouse().Y) {
		m.clickPlaylist(msg.Mouse().Y)
	} else if l.hasLists && l.entries.rows().contains(msg.Mouse().X, msg.Mouse().Y) {
		m.clickEntry(msg.Mouse().Y)
	} else if l.hasOutput && l.output.contains(msg.Mouse().X, msg.Mouse().Y) {
		m.session.Pane = paneOutput
	}
	return m, nil
}

// wheel scrolls the pane under the pointer.
func (m Model) wheel(l downloadLayout, msg tea.MouseMsg, dir int) (tea.Model, tea.Cmd) {
	// The wheel scrolls what is under the pointer and nothing else: the selection and the focus stay.
	switch {
	case l.hasLists && l.playlists.contains(msg.Mouse().X, msg.Mouse().Y):
		m.scrollPane(false, dir*pointer.WheelStep)
	case l.hasLists && l.entries.contains(msg.Mouse().X, msg.Mouse().Y):
		m.scrollPane(true, dir*pointer.WheelStep)
	case l.hasOutput && l.output.contains(msg.Mouse().X, msg.Mouse().Y):
		focus := m.session.Pane
		m.session.Pane = paneOutput
		m.move(dir * pointer.WheelStep)
		m.session.Pane = focus
	}
	return m, nil
}

func (m *Model) clickPlaylist(y int) {
	i, ok := PlaylistsAt(m.downloadProps(m.width, m.height), m.width, m.height, y)
	m.session.Pane = paneLists
	if !ok {
		return
	}
	m.session.Follow = false
	m.session.SelPlaylist, m.session.SelEntry, m.session.OutScroll = i, 0, 0
}

func (m *Model) clickEntry(y int) {
	m.session.Pane = paneEntries
	if m.session.SelPlaylist >= len(m.session.State.Playlists) {
		return
	}
	if pl := m.session.State.Playlists[m.session.SelPlaylist]; pl.Err != "" || !pl.Listed {
		return
	}
	i, ok := EntriesAt(m.downloadProps(m.width, m.height), m.width, m.height, y)
	if !ok {
		return
	}
	m.session.Follow = false
	m.session.SelEntry, m.session.OutScroll = i, 0
}

// PlaylistsBox is where the Playlists box is drawn, in the cells of the screen; ok is false
// when the screen draws no such box.
func (m Model) PlaylistsBox() (x, y, w, h int, ok bool) {
	l, drawn := m.layout()
	if !drawn || !l.hasLists {
		return 0, 0, 0, 0, false
	}
	return l.playlists.x, l.playlists.y, l.playlists.w, l.playlists.h, true
}

// WithReserve keeps rows rows under the playlists free, for the buttons the app draws there.
func (m Model) WithReserve(rows int) Model {
	m.reserve = rows
	return m
}

// LeftColumn is where the column of playlists is, in the cells of the screen, and how many
// of its bottom rows are kept for the app's buttons; ok is false when the screen has no
// such column.
func (m Model) LeftColumn() (x, y, w, h, reserve int, ok bool) {
	l, drawn := m.layout()
	if !drawn || !l.hasLists || !m.columns(m.width) {
		return 0, 0, 0, 0, 0, false
	}
	return l.playlists.x, l.playlists.y, l.playlists.w, l.playlists.h + m.reserve, m.reserve, true
}

// borderAt is the border between the links and the songs at column x: 1 for it, 0 for none. The cell each side of the line takes hold of it.
func (m Model) borderAt(x, y int) int {
	if !m.columns(m.width) || y < 3 || y >= m.height-1 {
		return 0
	}
	left := m.leftWidth(m.width)
	if x == left-1 || x == left {
		return 1
	}
	return 0
}

// dragBorder works the dragging of the borders between the columns: pressing on one takes
// hold of it, moving changes the width, letting go ends it. It says whether msg was meant
// for the border.
func (m Model) dragBorder(msg tea.MouseMsg) (Model, bool) {
	pos := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		if pos.Button == tea.MouseLeft {
			if b := m.borderAt(pos.X, pos.Y); b > 0 {
				m.dragging = b
				return m, true
			}
		}
	case tea.MouseMotionMsg:
		if m.dragging == 0 {
			return m, false
		}
		m.leftW = max(minPane, min(pos.X, m.width-minPane*2))
		return m, true
	case tea.MouseReleaseMsg:
		if m.dragging != 0 {
			m.dragging = 0
			return m, true
		}
	}
	return m, false
}
