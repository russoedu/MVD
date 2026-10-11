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
	if m.width < 60 || m.height < 16 || m.showHelp {
		return l, false
	}
	p := m.downloadProps(m.width, m.height)
	frame := Layout(p, m.width, m.height)

	box := func(name string) rect {
		r, _ := frame.RegionNamed(name)
		return rect{r.X, r.Y, r.W, r.H}
	}
	l.hasLists, l.hasOutput = p.LeftWidth > 0, m.fullLog || m.width >= minWideWidth
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
	if m.showHelp {
		if pointer.IsLeftClick(msg) {
			m.showHelp = false
		}
		return m, nil
	}
	if m.confirmQuit || m.adding {
		return m, nil
	}
	l, ok := m.layout()
	if !ok {
		return m, nil
	}

	if dir := pointer.WheelDirection(msg); dir != 0 {
		return m.wheel(l, msg, dir)
	}
	if !pointer.IsLeftClick(msg) {
		return m, nil
	}
	if msg.Mouse().Y == m.height-1 {
		if h, found := keybar.At(m.hints(), msg.Mouse().X); found {
			if key, clickable := keybar.KeyMsg(h.Key); clickable {
				return m.handleKey(key)
			}
		}
		return m, nil
	}
	if l.hasLists && l.playlists.rows().contains(msg.Mouse().X, msg.Mouse().Y) {
		m.clickPlaylist(msg.Mouse().Y)
	} else if l.hasLists && l.entries.rows().contains(msg.Mouse().X, msg.Mouse().Y) {
		m.clickEntry(msg.Mouse().Y)
	} else if l.hasOutput && l.output.contains(msg.Mouse().X, msg.Mouse().Y) {
		m.focus = paneOutput
	}
	return m, nil
}

// wheel scrolls the pane under the pointer.
func (m Model) wheel(l downloadLayout, msg tea.MouseMsg, dir int) (tea.Model, tea.Cmd) {
	switch {
	case l.hasLists && l.playlists.contains(msg.Mouse().X, msg.Mouse().Y):
		m.focus = paneLists
	case l.hasLists && l.entries.contains(msg.Mouse().X, msg.Mouse().Y):
		m.focus = paneEntries
	case l.hasOutput && l.output.contains(msg.Mouse().X, msg.Mouse().Y):
		m.focus = paneOutput
	default:
		return m, nil
	}
	m.move(dir * pointer.WheelStep)
	return m, nil
}

func (m *Model) clickPlaylist(y int) {
	i, ok := PlaylistsAt(m.downloadProps(m.width, m.height), m.width, m.height, y)
	m.focus = paneLists
	if !ok {
		return
	}
	m.follow = false
	m.selPlaylist, m.selEntry, m.outScroll = i, 0, 0
}

func (m *Model) clickEntry(y int) {
	m.focus = paneEntries
	if m.selPlaylist >= len(m.state.Playlists) {
		return
	}
	if pl := m.state.Playlists[m.selPlaylist]; pl.Err != "" || !pl.Listed {
		return
	}
	i, ok := EntriesAt(m.downloadProps(m.width, m.height), m.width, m.height, y)
	if !ok {
		return
	}
	m.follow = false
	m.selEntry, m.outScroll = i, 0
}
