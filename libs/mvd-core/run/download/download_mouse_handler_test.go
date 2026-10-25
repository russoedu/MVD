package download

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/pointer"
	"youtube-downloader/libs/mvd-core/testing/keys"
	"youtube-downloader/libs/mvd-core/testing/mouse"
)

func TestDownloadScreenClicksSelectPanesRowsAndKeyBarEntries(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	// Wide layout: the links box is columns 0-30 and the songs box the rest, with the details
	// under it; they start at row 3, the songs with a row of headings.

	next, _ := m.Update(mouse.Click(5, 5))
	m = next.(Model)
	if m.selPlaylist != 1 || m.focus != paneLists {
		t.Errorf("clicking the second playlist: selPlaylist %d, focus %d", m.selPlaylist, m.focus)
	}

	m.selPlaylist = 0
	next, _ = m.Update(mouse.Click(40, 7)) // the heading row is at row 4, so the rows start at row 5
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
	m.technical = true

	next, _ := m.Update(mouse.Wheel(90, 30, tea.MouseWheelUp)) // over the output box, under the songs
	m = next.(Model)
	if m.focus != paneLists || m.outScroll != pointer.WheelStep {
		t.Errorf("wheel up over the output: focus %d, scrolled %d", m.focus, m.outScroll)
	}

	next, _ = m.Update(mouse.Wheel(40, 6, tea.MouseWheelDown)) // over the songs
	m = next.(Model)
	if m.focus != paneLists || m.selEntry != 4 {
		t.Errorf("the wheel over the songs moved the selection: focus %d, selEntry %d", m.focus, m.selEntry)
	}
}

func TestTheWheelScrollsTheSongsAndAClickDoesNotMoveThem(t *testing.T) {
	m := sampleModel(t)
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Many", Entries: manyEntries(60)})
	m.width, m.height = 120, 30
	m.focus, m.selEntry = paneEntries, 0

	next, _ := m.Update(mouse.Wheel(40, 8, tea.MouseWheelDown))
	m = next.(Model)
	if m.entriesTop != pointer.WheelStep || m.selEntry != 0 {
		t.Fatalf("after the wheel: top %d, selection %d", m.entriesTop, m.selEntry)
	}
	top := m.entriesTop
	next, _ = m.Update(mouse.Click(40, 12)) // a song in view
	m = next.(Model)
	if m.entriesTop != top || m.selEntry == 0 {
		t.Errorf("a click moved the songs: top %d (was %d), selection %d", m.entriesTop, top, m.selEntry)
	}
	for i := 0; i < 40; i++ { // the keys go past the bottom: it scrolls just enough
		next, _ = m.Update(keys.Press("down"))
		m = next.(Model)
	}
	l, _ := m.layout()
	if h := l.entries.rows().h - 1; m.selEntry < m.entriesTop || m.selEntry >= m.entriesTop+h {
		t.Errorf("the selection %d is out of view: top %d, %d rows", m.selEntry, m.entriesTop, h)
	}
}

// manyEntries are n songs for a playlist.
func manyEntries(n int) []engine.EntryInfo {
	out := make([]engine.EntryInfo, n)
	for i := range out {
		out[i] = engine.EntryInfo{ID: 100 + i, Playlist: 0, Index: i + 1, VideoID: "v", Title: "song"}
	}

	return out
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

func TestTheBorderBetweenTheLinksAndTheSongsCanBeDragged(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	left := m.leftWidth(120)

	// Take hold of the border after the links and move it ten cells right.
	next, _ := m.Update(mouse.Click(left-1, 10))
	m = next.(Model)
	if m.dragging != 1 {
		t.Fatalf("dragging %d", m.dragging)
	}
	next, _ = m.Update(tea.MouseMotionMsg{X: left + 10, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(tea.MouseReleaseMsg{X: left + 10, Y: 10, Button: tea.MouseLeft})
	m = next.(Model)
	if m.dragging != 0 || m.leftWidth(120) != left+10 {
		t.Errorf("after the drag: left %d (was %d), dragging %d", m.leftWidth(120), left, m.dragging)
	}

	// The songs take the rest of the width: there is no other border to hold.
	if m.borderAt(m.width/2, 10) != 0 {
		t.Error("a border was found in the middle of the songs")
	}
}
