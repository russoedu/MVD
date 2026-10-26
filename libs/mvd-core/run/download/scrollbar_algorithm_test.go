package download

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
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
		if tr.entries {
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
