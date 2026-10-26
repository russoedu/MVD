package download

import (
	"strings"

	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// scrollTrack is a scrollbar's place: its column, where its track begins, how long it is, and the
// rows it is for, scrolled down by top.
type scrollTrack struct {
	x, y, h int
	n, top  int
	entries bool
}

// scrollTracks are the scrollbars the panes need now: one on the right edge of each pane whose rows
// do not all show.
func (m Model) scrollTracks(l downloadLayout) []scrollTrack {
	if !l.hasLists {
		return nil
	}
	var tracks []scrollTrack
	pl := l.playlists.rows()
	if n := len(m.session.State.Playlists); n > pl.h {
		tracks = append(tracks, scrollTrack{x: l.playlists.x + l.playlists.w - 1, y: pl.y, h: pl.h, n: n, top: m.playlistsTop})
	}
	if m.tableShown() {
		en := l.entries.rows()
		if n := len(m.session.VisibleEntries()); n > en.h-1 {
			tracks = append(tracks, scrollTrack{x: l.entries.x + l.entries.w - 1, y: en.y + 1, h: en.h - 1, n: n, top: m.entriesTop, entries: true})
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

// scrollbarClick scrolls the songs to the place of the bar that was clicked, and says whether the
// click was on that bar.
func (m Model) scrollbarClick(l downloadLayout, x, y int) (Model, bool) {
	for _, t := range m.scrollTracks(l) {
		if !t.entries || x != t.x || y < t.y || y >= t.y+t.h {
			continue
		}
		m.entriesTop = topAt(t.n, t.h, y-t.y)

		return m, true
	}

	return m, false
}
