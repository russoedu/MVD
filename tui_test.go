package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// sampleModel builds a model with a few playlists and entries in every state.
func sampleModel(t *testing.T) tuiModel {
	t.Helper()
	eng, err := NewEngine(defaultConfig(), "yt-dlp", []string{"https://a", "https://b", "https://c"}, "")
	if err != nil {
		t.Fatal(err)
	}
	m := newTUIModel(eng)
	m.state.apply(EvPlaylistListed{Playlist: 0, Title: "90s UK Dance Hits", Entries: []EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "rnlp_avexYQ", Title: "You Don't Know Me (feat. Duane Harden) (Radio Edit)", Channel: "Armand Van Helden - Topic"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "LIWrcD4ImwA", Title: "Don't Call Me Baby", Channel: "Madison Avenue - Topic"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "MR3uP7IYz44", Title: "Modjo - Lady (Hear Me Tonight)", Channel: "ModjoOfficial"},
		{ID: 3, Playlist: 0, Index: 4, VideoID: "GUypoENfCtY", Title: "Bob Sinclar - I Feel For You", Channel: "Bob Sinclar"},
		{ID: 4, Playlist: 0, Index: 5, VideoID: "hRvrj_diWYQ", Title: "Stardust - Music Sounds Better With You 日本語タイトル", Channel: "Stardust"},
		{ID: 5, Playlist: 0, Index: 6, VideoID: "X4UGPCR3zEQ", Title: "The Bucketheads - The Bomb", Channel: "The Bucketheads - Topic"},
	}})
	m.state.apply(EvPlaylistListed{Playlist: 1, Title: "Ibiza Classics", Entries: []EntryInfo{
		{ID: 6, Playlist: 1, Index: 1, VideoID: "Jb6gcoR266U", Title: "Daft Punk - One More Time", Channel: "Daft Punk"},
	}})
	m.state.apply(EvPlaylistFailed{Playlist: 2, Err: "yt-dlp --flat-playlist failed: ERROR: unable to recognize playlist"})
	m.state.apply(EvEntryState{Entry: 0, State: StateDone, TargetID: "-bsONE-kZwI", Official: true})
	m.state.apply(EvEntryState{Entry: 1, State: StateDuplicate, TargetID: "-bsONE-kZwI"})
	m.state.apply(EvEntryState{Entry: 2, State: StateFailed, TargetID: "MR3uP7IYz44", Err: "Video unavailable"})
	m.state.apply(EvEntryState{Entry: 3, State: StateResolving, TargetID: "GUypoENfCtY"})
	m.state.apply(EvEntryState{Entry: 4, State: StateDownloading, TargetID: "hRvrj_diWYQ"})
	m.state.apply(EvProgress{Entry: 4, Percent: 34.2, Downloaded: 40_000_000, Total: 117_833_728, Speed: 8.1 * 1024 * 1024, ETA: 9})
	m.state.apply(EvEntryState{Entry: 5, State: StateMerging, TargetID: "X4UGPCR3zEQ"})
	for i := 0; i < 50; i++ {
		m.state.apply(EvLog{Playlist: 0, Entry: 4, Line: "[download] line " + strings.Repeat("x", i)})
	}
	m.state.apply(EvLog{Playlist: 0, Entry: -1, Line: "listing https://a"})
	return m
}

func checkFrame(t *testing.T, name, frame string, w, h int) {
	t.Helper()
	lines := strings.Split(frame, "\n")
	if len(lines) != h {
		t.Errorf("%s: want %d lines, got %d", name, h, len(lines))
	}
	for i, l := range lines {
		if lw := ansi.StringWidth(l); lw != w {
			t.Errorf("%s: line %d has width %d, want %d: %q", name, i, lw, w, ansi.Strip(l))
		}
	}
}

func TestViewLayouts(t *testing.T) {
	sizes := [][2]int{{120, 36}, {100, 24}, {80, 30}, {60, 16}, {200, 60}}
	for _, sz := range sizes {
		m := sampleModel(t)
		m.width, m.height = sz[0], sz[1]
		m.focus = paneEntries
		m.selEntry = 4
		checkFrame(t, "entries", m.View(), sz[0], sz[1])

		m.focus = paneLists
		checkFrame(t, "lists", m.View(), sz[0], sz[1])

		m.fullLog = true
		m.focus = paneOutput
		m.outScroll = 5
		checkFrame(t, "fulllog", m.View(), sz[0], sz[1])

		m.fullLog = false
		m.filterFailed = true
		m.focus = paneEntries
		checkFrame(t, "filter", m.View(), sz[0], sz[1])

		m.confirmQuit = true
		checkFrame(t, "confirm", m.View(), sz[0], sz[1])

		m.confirmQuit = false
		m.showHelp = true
		checkFrame(t, "help", m.View(), sz[0], sz[1])
	}
}

func TestViewContent(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.focus = paneEntries
	m.selEntry = 4
	frame := ansi.Strip(m.View())
	for _, want := range []string{"Queue 1", "Running 3", "Done 1", "Official 1", "Dup 1", "Failed 1", "90s UK Dance Hits", "Ibiza Classics", "3/6", "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09", "ERR", "dup", "res.", "merge", "⇄", "follow"} {
		if !strings.Contains(frame, want) {
			t.Errorf("frame missing %q\n%s", want, frame)
		}
	}
	// Failed playlist shows its error in the entries pane.
	m.selPlaylist = 2
	m.focus = paneLists
	if f := ansi.Strip(m.View()); !strings.Contains(f, "unable to recognize playlist") {
		t.Errorf("playlist error not shown\n%s", f)
	}
}

func TestKeyHandling(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36

	press := func(k string) {
		var mm tea.Model
		var msg tea.KeyMsg
		switch k {
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		mm, _ = m.Update(msg)
		m = mm.(tuiModel)
	}

	if m.focus != paneLists {
		t.Fatal("initial focus should be playlists")
	}
	press("tab")
	if m.focus != paneEntries {
		t.Fatal("tab should move to entries")
	}
	press("down")
	press("down")
	if m.selEntry != 2 || m.follow {
		t.Fatalf("manual navigation should move selection and stop following: sel=%d follow=%v", m.selEntry, m.follow)
	}
	press("f")
	if !m.filterFailed || len(m.visibleEntries()) != 1 {
		t.Fatalf("filter should show the single failed entry, got %d", len(m.visibleEntries()))
	}
	press("f")
	press("enter")
	if !m.follow || m.selectedEntry().ID != 3 {
		t.Fatalf("follow should jump to the active entry, got %d follow=%v", m.selectedEntry().ID, m.follow)
	}
	press("l")
	if !m.fullLog || m.focus != paneOutput {
		t.Fatal("l should open the full log")
	}
	press("l")
	press("q")
	if !m.confirmQuit {
		t.Fatal("q while running should ask for confirmation")
	}
	press("n")
	if m.confirmQuit {
		t.Fatal("n should cancel quitting")
	}
	press("?")
	if !m.showHelp {
		t.Fatal("? should show help")
	}
	press("x")
	if m.showHelp {
		t.Fatal("any key should close help")
	}

	// A new download jumps the selection while following.
	m.follow = true
	mm, _ := m.Update(evMsg{EvEntryState{Entry: 6, State: StateDownloading, TargetID: "Jb6gcoR266U"}})
	m = mm.(tuiModel)
	if m.selPlaylist != 1 || m.selectedEntry().ID != 6 {
		t.Fatalf("follow should jump to playlist 1 entry 6, got playlist %d entry %d", m.selPlaylist, m.selectedEntry().ID)
	}

	// Once idle, q quits straight away.
	m.state.Idle = true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q when idle should quit")
	}
}

func TestRenderBoxAndWindow(t *testing.T) {
	box := renderBox("Title", "right", 30, 5, []string{"one", strings.Repeat("y", 40)}, true)
	checkFrame(t, "box", box, 30, 5)
	plain := ansi.Strip(box)
	if !strings.Contains(plain, "╭─ Title ") || !strings.Contains(plain, " right ─╮") || !strings.Contains(plain, "…") {
		t.Errorf("unexpected box:\n%s", plain)
	}

	rows := []string{"a", "b", "c", "d", "e", "f"}
	if got := window(rows, 5, 3); strings.Join(got, "") != "def" {
		t.Errorf("window at end wrong: %v", got)
	}
	if got := window(rows, 0, 3); strings.Join(got, "") != "abc" {
		t.Errorf("window at start wrong: %v", got)
	}
	if got := window(rows, 3, 3); strings.Join(got, "") != "cde" {
		t.Errorf("window in middle wrong: %v", got)
	}
}
