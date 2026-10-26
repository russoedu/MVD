package download

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
	"youtube-downloader/libs/mvd-core/pointer"
)

// barKind says which pane a scrollbar belongs to.
type barKind int

const (
	barNone barKind = iota
	barLinks
	barSongs
	barLog
)

// scrollTrack is a scrollbar's place: its column, where its track begins, how long it is, and the
// rows it is for, with the first one that shows.
type scrollTrack struct {
	kind    barKind
	x, y, h int
	n, top  int
}

// scrollTracks are the scrollbars the panes need now: one on the right edge of each pane whose rows
// do not all show.
func (m Model) scrollTracks(l downloadLayout) []scrollTrack {
	var tracks []scrollTrack
	if l.hasLists {
		pl := l.playlists.rows()
		if n := len(m.session.State.Playlists); n > pl.h {
			tracks = append(tracks, scrollTrack{kind: barLinks, x: l.playlists.x + l.playlists.w - 1, y: pl.y, h: pl.h, n: n, top: m.playlistsTop})
		}
		if m.tableShown() {
			en := l.entries.rows()
			if n := len(m.session.VisibleEntries()); n > en.h-1 {
				tracks = append(tracks, scrollTrack{kind: barSongs, x: l.entries.x + l.entries.w - 1, y: en.y + 1, h: en.h - 1, n: n, top: m.entriesTop})
			}
		}
	}
	if l.hasOutput && m.session.Technical {
		out := l.output.rows()
		h := out.h
		if len(m.outputFooter().lines) > 0 {
			h -= 2
		}
		if n := len(m.session.OutputLines()); n > h && h > 0 {
			// The log is scrolled up from its bottom: the first row that shows is n-h-OutScroll.
			tracks = append(tracks, scrollTrack{kind: barLog, x: l.output.x + l.output.w - 1, y: out.y, h: h, n: n, top: n - h - m.session.OutScroll})
		}
	}

	return tracks
}

// withScrollbars draws the bar of each pane that scrolls over the right edge of its box, with an arrow
// at an end that has more beyond it.
func (m Model) withScrollbars(screen string) string {
	if m.session.ShowHelp || m.session.Adding || m.session.Swapping {
		return screen
	}
	l, ok := m.layout()
	if !ok {
		return screen
	}
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(look.Colours.Focus)).Bold(true)
	arrow := lipgloss.NewStyle().Foreground(lipgloss.Color(look.Colours.Highlight))
	for _, t := range m.scrollTracks(l) {
		start, length, scrolls := thumb(t.n, t.h, t.top)
		if !scrolls {
			continue
		}
		cells := make([]string, t.h)
		for i := range cells {
			switch {
			case i >= start && i < start+length:
				cells[i] = bar.Render("┃")
			default:
				cells[i] = look.Dim.Render("│")
			}
		}
		if start > 0 {
			cells[0] = arrow.Render("↑")
		}
		if start+length < t.h {
			cells[t.h-1] = arrow.Render("↓")
		}
		screen = dialog.Overlay(screen, strings.Join(cells, "\n"), t.x, t.y)
	}

	return screen
}

// scrolledBy takes the pane of a bar to the place of the track the pointer is at.
func (m Model) scrolledBy(t scrollTrack, y int) Model {
	top := topAt(t.n, t.h, min(max(y-t.y, 0), t.h-1))
	switch t.kind {
	case barLinks:
		m.playlistsTop = top
	case barSongs:
		m.entriesTop = top
	case barLog:
		m.session.OutScroll = max(t.n-t.h-top, 0)
	}

	return m
}

// scrollbarMouse works the mouse on the scrollbars: a click on a track scrolls there and takes hold of
// the bar, which then follows the pointer until it is let go. It says whether msg was for a bar.
func (m Model) scrollbarMouse(l downloadLayout, msg tea.MouseMsg) (Model, bool) {
	pos := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		if !pointer.IsLeftClick(msg) {
			return m, false
		}
		for _, t := range m.scrollTracks(l) {
			if pos.X == t.x && pos.Y >= t.y && pos.Y < t.y+t.h {
				m.barDrag = t.kind

				return m.scrolledBy(t, pos.Y), true
			}
		}
	case tea.MouseMotionMsg:
		if m.barDrag == barNone {
			return m, false
		}
		for _, t := range m.scrollTracks(l) {
			if t.kind == m.barDrag {
				return m.scrolledBy(t, pos.Y), true
			}
		}
		m.barDrag = barNone

		return m, true
	case tea.MouseReleaseMsg:
		if m.barDrag != barNone {
			m.barDrag = barNone

			return m, true
		}
	}

	return m, false
}
