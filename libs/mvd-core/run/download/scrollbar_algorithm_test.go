package download

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/testing/mouse"
)

func TestThereIsNoBarWhenEveryRowShows(t *testing.T) {
	if _, _, ok := thumb(5, 5, 0); ok {
		t.Error("a bar for rows that all show")
	}
}

func TestTheBarRunsFromTheTopToTheBottomAsThePaneScrolls(t *testing.T) {
	n, h := 100, 10
	if start, length, _ := thumb(n, h, 0); start != 0 || length != 1 {
		t.Errorf("at the top: %d %d", start, length)
	}
	if start, length, _ := thumb(n, h, n-h); start+length != h {
		t.Errorf("at the bottom the bar ends at %d of %d", start+length, h)
	}
	if start, _, _ := thumb(n, h, 1000); start != h-1 {
		t.Errorf("past the end it stays on the track: %d", start)
	}
}

func TestTheBarIsLongerTheMoreOfTheListShows(t *testing.T) {
	_, short, _ := thumb(20, 10, 0)
	_, long, _ := thumb(200, 10, 0)
	if short <= long {
		t.Errorf("half the list shows a bar of %d, a twentieth one of %d", short, long)
	}
}

func TestAClickOnTheTrackScrollsToThatPlace(t *testing.T) {
	n, h := 100, 10
	if got := topAt(n, h, 0); got != 0 {
		t.Errorf("the top of the track: %d", got)
	}
	if got := topAt(n, h, h-1); got != n-h {
		t.Errorf("the bottom of the track: %d", got)
	}
	if mid := topAt(n, h, h/2); mid < 30 || mid > 60 {
		t.Errorf("the middle of the track: %d", mid)
	}
}

func TestTheSongsShowABarThatFollowsTheWheelAndTakesAClick(t *testing.T) {
	m := sampleModel(t)
	m.session.State.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Many", Entries: manyEntries(60)})
	m.width, m.height = 120, 30
	m.session.Pane = paneEntries
	if !strings.Contains(ansi.Strip(m.render()), "┃") {
		t.Fatal("sixty songs scroll, and no bar is drawn")
	}
	l, _ := m.layout()
	tracks := m.scrollTracks(l)
	var songs scrollTrack
	for _, tr := range tracks {
		if tr.kind == barSongs {
			songs = tr
		}
	}
	next, _ := m.Update(mouse.Click(songs.x, songs.y+songs.h-1))
	m = next.(Model)
	if m.entriesTop != songs.n-songs.h {
		t.Errorf("a click at the bottom of the bar: top %d, want %d", m.entriesTop, songs.n-songs.h)
	}
	if !strings.Contains(ansi.Strip(m.render()), "↑") {
		t.Error("scrolled to the bottom, no arrow says there is more above")
	}
}

func trackOf(m Model, kind barKind) (scrollTrack, bool) {
	l, _ := m.layout()
	for _, tr := range m.scrollTracks(l) {
		if tr.kind == kind {
			return tr, true
		}
	}

	return scrollTrack{}, false
}

func TestTheBarCanBeDraggedAndLetGo(t *testing.T) {
	m := sampleModel(t)
	m.session.State.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "Many", Entries: manyEntries(60)})
	m.width, m.height = 120, 30
	tr, _ := trackOf(m, barSongs)
	next, _ := m.Update(mouse.Click(tr.x, tr.y))
	m = next.(Model)
	if m.barDrag != barSongs || m.entriesTop != 0 {
		t.Fatalf("pressing on the top of the bar: drag %d, top %d", m.barDrag, m.entriesTop)
	}
	next, _ = m.Update(tea.MouseMotionMsg{X: tr.x, Y: tr.y + tr.h - 1, Button: tea.MouseLeft})
	m = next.(Model)
	if m.entriesTop != tr.n-tr.h {
		t.Errorf("dragged to the bottom: top %d, want %d", m.entriesTop, tr.n-tr.h)
	}
	next, _ = m.Update(tea.MouseReleaseMsg{X: tr.x, Y: tr.y, Button: tea.MouseLeft})
	m = next.(Model)
	if m.barDrag != barNone {
		t.Error("the bar is still held after the button was let go")
	}
}

func TestTheLinksHaveABarToo(t *testing.T) {
	m := sampleModel(t)
	sources := make([]engine.PlaylistSource, 30)
	for i := range sources {
		sources[i] = engine.PlaylistSource{Index: i, URL: "https://youtu.be/x"}
	}
	m.session.State = runstate.New(sources)
	m.width, m.height = 120, 30
	tr, ok := trackOf(m, barLinks)
	if !ok {
		t.Fatal("thirty links scroll, and there is no bar")
	}
	next, _ := m.Update(mouse.Click(tr.x, tr.y+tr.h-1))
	m = next.(Model)
	if m.playlistsTop != tr.n-tr.h || m.dragging != 0 {
		t.Errorf("a click at the bottom of the bar: top %d, want %d, border dragging %d", m.playlistsTop, tr.n-tr.h, m.dragging)
	}
}

func TestTheLogHasABarThatScrollsItBackwards(t *testing.T) {
	m := sampleModel(t)
	m.width, m.height = 120, 36
	m.session.SelEntry, m.session.Technical = 4, true // the song with a long log
	tr, ok := trackOf(m, barLog)
	if !ok {
		t.Skip("the log of the sample song fits its pane")
	}
	next, _ := m.Update(mouse.Click(tr.x, tr.y))
	m = next.(Model)
	if m.session.OutScroll != tr.n-tr.h {
		t.Errorf("the top of the bar shows the oldest lines: scrolled %d, want %d", m.session.OutScroll, tr.n-tr.h)
	}
}
