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
	// Wide layout: the playlists box fills rows 3-7 (its rows 4-6), the entries box starts at row 8
	// (its rows 9 and below), and the output box starts at column 45.

	next, _ := m.Update(mouse.Click(5, 5))
	m = next.(Model)
	if m.selPlaylist != 1 || m.focus != paneLists {
		t.Errorf("clicking the second playlist: selPlaylist %d, focus %d", m.selPlaylist, m.focus)
	}

	m.selPlaylist = 0
	next, _ = m.Update(mouse.Click(5, 11))
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

	next, _ := m.Update(mouse.Wheel(80, 10, tea.MouseWheelUp)) // over the output box
	m = next.(Model)
	if m.focus != paneOutput || m.outScroll != pointer.WheelStep {
		t.Errorf("wheel up over the output: focus %d, scrolled %d", m.focus, m.outScroll)
	}

	next, _ = m.Update(mouse.Wheel(5, 11, tea.MouseWheelDown)) // over the entries
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
