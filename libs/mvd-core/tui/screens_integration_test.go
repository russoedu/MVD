package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
)

// fakeController stands in for the engine.
type fakeController struct {
	events  chan interface{}
	sources []engine.PlaylistSource
	retried []int
}

func (f *fakeController) Events() <-chan interface{}       { return f.events }
func (f *fakeController) Sources() []engine.PlaylistSource { return f.sources }
func (f *fakeController) LogPath() string                  { return "mvd.log" }
func (f *fakeController) Retry(id int) bool                { f.retried = append(f.retried, id); return true }
func (f *fakeController) RetryPlaylist(int) int            { return 2 }

// sampleModel builds a model with a few playlists and entries in every state.
func sampleModel(t *testing.T) model {
	t.Helper()
	c := &fakeController{events: make(chan interface{}), sources: []engine.PlaylistSource{{Index: 0, URL: "https://a"}, {Index: 1, URL: "https://b"}, {Index: 2, URL: "https://c"}}}
	m := newModel(c)
	m.state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "90s UK Dance Hits", Entries: []engine.EntryInfo{
		{ID: 0, Playlist: 0, Index: 1, VideoID: "rnlp_avexYQ", Title: "You Don't Know Me (feat. Duane Harden) (Radio Edit)", Channel: "Armand Van Helden - Topic"},
		{ID: 1, Playlist: 0, Index: 2, VideoID: "LIWrcD4ImwA", Title: "Don't Call Me Baby", Channel: "Madison Avenue - Topic"},
		{ID: 2, Playlist: 0, Index: 3, VideoID: "MR3uP7IYz44", Title: "Modjo - Lady (Hear Me Tonight)", Channel: "ModjoOfficial"},
		{ID: 3, Playlist: 0, Index: 4, VideoID: "GUypoENfCtY", Title: "Bob Sinclar - I Feel For You", Channel: "Bob Sinclar"},
		{ID: 4, Playlist: 0, Index: 5, VideoID: "hRvrj_diWYQ", Title: "Stardust - Music Sounds Better With You 日本語タイトル 🔊", Channel: "Stardust"},
		{ID: 5, Playlist: 0, Index: 6, VideoID: "X4UGPCR3zEQ", Title: "", Channel: "The Bucketheads - Topic"},
	}})
	m.state.Apply(engine.EvPlaylistListed{Playlist: 1, Title: "Música Eletrônica 1990s 🔊 1999, 1998, 1997, 1996, 1995, 1994, 1993, 1992, 1991, 1990", Entries: []engine.EntryInfo{
		{ID: 6, Playlist: 1, Index: 1, VideoID: "Jb6gcoR266U", Title: "Daft Punk - One More Time", Channel: "Daft Punk"},
	}})
	m.state.Apply(engine.EvPlaylistFailed{Playlist: 2, Err: "yt-dlp --flat-playlist failed: ERROR: unable to recognize playlist"})
	m.state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone, TargetID: "-bsONE-kZwI", Official: true})
	m.state.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDuplicate, TargetID: "-bsONE-kZwI"})
	m.state.Apply(engine.EvEntryState{Entry: 2, State: engine.StateFailed, TargetID: "MR3uP7IYz44", Err: "Video unavailable"})
	m.state.Apply(engine.EvEntryState{Entry: 3, State: engine.StateResolving, TargetID: "GUypoENfCtY"})
	m.state.Apply(engine.EvEntryState{Entry: 4, State: engine.StateDownloading, TargetID: "hRvrj_diWYQ"})
	m.state.Apply(engine.EvProgress{Entry: 4, Percent: 34.2, Downloaded: 40_000_000, Total: 117_833_728, Speed: 8.1 * 1024 * 1024, ETA: 9})
	m.state.Apply(engine.EvEntryState{Entry: 5, State: engine.StateMerging, TargetID: "X4UGPCR3zEQ"})
	for i := 0; i < 50; i++ {
		m.state.Apply(engine.EvLog{Playlist: 0, Entry: 4, Line: "[download] line\t" + strings.Repeat("x", i)})
	}
	m.state.Apply(engine.EvLog{Playlist: 0, Entry: -1, Line: "listing https://a"})
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
		checkFrame(t, "entries", m.View().Content, sz[0], sz[1])

		m.focus = paneLists
		checkFrame(t, "lists", m.View().Content, sz[0], sz[1])

		m.selPlaylist = 1
		checkFrame(t, "emoji playlist", m.View().Content, sz[0], sz[1])
		m.selPlaylist = 0

		m.fullLog = true
		m.focus = paneOutput
		m.outScroll = 5
		checkFrame(t, "fulllog", m.View().Content, sz[0], sz[1])

		m.fullLog = false
		m.filterFailed = true
		m.focus = paneEntries
		checkFrame(t, "filter", m.View().Content, sz[0], sz[1])

		m.confirmQuit = true
		checkFrame(t, "confirm", m.View().Content, sz[0], sz[1])

		m.confirmQuit = false
		m.showHelp = true
		checkFrame(t, "help", m.View().Content, sz[0], sz[1])
	}
}

func TestViewContent(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.focus = paneEntries
	m.selEntry = 4
	frame := ansi.Strip(m.View().Content)
	for _, want := range []string{"Queue 1", "Running 3", "Done 1", "Official 1", "Dup 1", "Failed 1", "90s UK Dance Hits", "Música Eletrônica 1990s  1999", "3/6", "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09", "ERR", "dup", "res.", "merge", "⇄", "follow", "(X4UGPCR3zEQ)"} {
		if !strings.Contains(frame, want) {
			t.Errorf("frame missing %q\n%s", want, frame)
		}
	}
	if strings.Contains(frame, "🔊") || strings.Contains(frame, "\t") {
		t.Errorf("emoji and tabs must not reach the screen\n%s", frame)
	}
	m.selPlaylist = 2
	m.focus = paneLists
	if f := ansi.Strip(m.View().Content); !strings.Contains(f, "unable to recognize playlist") {
		t.Errorf("playlist error not shown\n%s", f)
	}
}

func TestKeyHandling(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36

	press := func(k string) {
		var mm tea.Model
		var msg tea.KeyPressMsg
		switch k {
		case "tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		case "down":
			msg = tea.KeyPressMsg{Code: tea.KeyDown}
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		default:
			msg = keyText(k)
		}
		mm, _ = m.Update(msg)
		m = mm.(model)
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
	press("r")
	if got := m.ctrl.(*fakeController).retried; len(got) != 1 || got[0] != 2 {
		t.Fatalf("r on a failed entry should retry it, got %v", got)
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

	m.follow = true
	mm, _ := m.Update(evMsg{engine.EvEntryState{Entry: 6, State: engine.StateDownloading, TargetID: "Jb6gcoR266U"}})
	m = mm.(model)
	if m.selPlaylist != 1 || m.selectedEntry().ID != 6 {
		t.Fatalf("follow should jump to playlist 1 entry 6, got playlist %d entry %d", m.selPlaylist, m.selectedEntry().ID)
	}

	m.state.Idle = true
	_, cmd := m.Update(keyText("q"))
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
