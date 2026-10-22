package download

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/testing/mouse"
)

func TestDownloadScreenClicksSelectPanesRowsAndKeyBarEntries(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	// Wide layout, in columns: the playlists box is columns 0-30, the entries box columns
	// 31-66 and the output box starts at column 67; all start at row 3, their rows at row 4.

	next, _ := m.Update(mouse.Click(5, 5))
	m = next.(Model)
	if m.selPlaylist != 1 || m.focus != paneLists {
		t.Errorf("clicking the second playlist: selPlaylist %d, focus %d", m.selPlaylist, m.focus)
	}

	m.selPlaylist = 0
	next, _ = m.Update(mouse.Click(40, 6))
	m = next.(Model)
	if m.selEntry != 2 || m.focus != paneEntries || m.follow {
		t.Errorf("clicking the third entry: selEntry %d, focus %d, follow %v", m.selEntry, m.focus, m.follow)
	}

	m.focus = paneLists
	next, _ = m.Update(mouse.Click(11, 35)) // "tab pane" in the key bar
	m = next.(Model)
	if m.focus != paneEntries {
		t.Errorf("clicking tab in the key bar should change pane, focus %d", m.focus)
	}
}

func TestDownloadScreenWheelScrollsThePaneUnderThePointer(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.selEntry = 4 // the entry with a long log

	next, _ := m.Update(mouse.Wheel(90, 10, tea.MouseWheelUp)) // over the output box
	m = next.(Model)
	if m.focus != paneOutput || m.outScroll != pointer.WheelStep {
		t.Errorf("wheel up over the output: focus %d, scrolled %d", m.focus, m.outScroll)
	}

	next, _ = m.Update(mouse.Wheel(40, 6, tea.MouseWheelDown)) // over the entries
	m = next.(Model)
	if m.focus != paneEntries || m.selEntry != 5 {
		t.Errorf("wheel down over the entries: focus %d, selEntry %d", m.focus, m.selEntry)
	}
}

func TestDownloadScreenIgnoresTheMouseWhileConfirmingToQuit(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.confirmQuit = true

	next, _ := m.Update(mouse.Click(5, 5))
	if next.(Model).selPlaylist != m.selPlaylist {
		t.Error("a click while confirming should do nothing")
	}
}

func TestTheBordersBetweenTheColumnsCanBeDragged(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	left, mid := m.leftWidth(120), m.midWidth(120)

	// Take hold of the border after the playlists and move it ten cells right.
	next, _ := m.Update(mouse.Click(left-1, 10))
	m = next.(Model)
	if m.dragging != 1 {
		t.Fatalf("dragging %d", m.dragging)
	}
	next, _ = m.Update(tea.MouseMotionMsg{X: left + 10, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(tea.MouseReleaseMsg{X: left + 10, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	if m.dragging != 0 || m.leftWidth(120) != left+10 || m.midWidth(120) != mid {
		t.Errorf("after the drag: left %d (was %d), mid %d (was %d), dragging %d", m.leftWidth(120), left, m.midWidth(120), mid, m.dragging)
	}

	// The border after the entries changes only the width of the entries.
	edge := m.leftWidth(120) + m.midWidth(120)
	next, _ = m.Update(mouse.Click(edge, 10))
	m = next.(Model)
	next, _ = m.Update(tea.MouseMotionMsg{X: edge - 6, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(tea.MouseReleaseMsg{X: edge - 6, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	if m.leftWidth(120) != left+10 || m.midWidth(120) != mid-6 {
		t.Errorf("after the second drag: left %d, mid %d", m.leftWidth(120), m.midWidth(120))
	}
}
