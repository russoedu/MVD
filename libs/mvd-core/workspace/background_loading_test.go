package workspace

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"testing"
	downloadpage "youtube-downloader/libs/mvd-core/run/download"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
)

// addingRun is a fake run that takes more links, and remembers them.
type addingRun struct {
	*fakeRun
	added []string
}

func (a *addingRun) AddSource(url string) (engine.PlaylistSource, bool) {
	a.added = append(a.added, url)

	return engine.PlaylistSource{Index: len(a.added), URL: url}, true
}

func autoLoadHost(t *testing.T, run ListRun) Host {
	t.Helper()
	host, _ := newListHost(t, newFakeRun())
	host.AutoLoad = true
	host.Start = func(config.Config, []string, []engine.PlannedEntry, bool) (ListRun, error) { return run, nil }

	return host
}

func TestALinkIsLoadedInTheBackgroundAsSoonAsItIsThere(t *testing.T) {
	m := sized(New(autoLoadHost(t, newFakeRun()), config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	if !m.starting {
		t.Fatal("a list with a link did not begin to load by itself")
	}
	empty := sized(New(autoLoadHost(t, newFakeRun()), config.Config{OutputDir: t.TempDir()}, downloadlist.New(nil)))
	if empty.starting {
		t.Error("an empty list began to load")
	}
}

func TestAListThatWasStoppedIsNotLoadedAgainUntilItChanges(t *testing.T) {
	fake := newFakeRun()
	m := sized(New(autoLoadHost(t, fake), config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m = m.runStartedFor(t, fake)
	m, _ = m.stop(false)
	next, _ := m.Update(runClosedMsg{})
	m = next.(Model)
	if m.starting || m.running() {
		t.Error("a stopped list began to load again by itself")
	}
}

func TestALinkAddedWhileLoadingGoesToTheRun(t *testing.T) {
	run := &addingRun{fakeRun: newFakeRun()}
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m = m.runStartedFor(t, run)

	m, _ = choose(m, entryAddLink)
	if !m.addOpen {
		t.Fatal("the window for adding did not open")
	}
	m = typeInto(m, "https://youtu.be/second")
	m, _ = press(m, "enter")
	if m.addOpen || len(run.added) != 1 || run.added[0] != "https://youtu.be/second" {
		t.Errorf("window open %v, added %v", m.addOpen, run.added)
	}
	if got := addressesOf(m); len(got) != 2 {
		t.Errorf("the list has %v", got)
	}
}

func (m Model) runStartedFor(t *testing.T, run ListRun) Model {
	t.Helper()
	next, _ := m.Update(runStartedMsg{run: run})

	return next.(Model)
}

// twoLinkRun is a fake run with two links, that remembers which was taken out.
type twoLinkRun struct {
	*addingRun
	removed []int
}

func (r *twoLinkRun) Sources() []engine.PlaylistSource {
	return []engine.PlaylistSource{{Index: 0, URL: "https://youtu.be/a"}, {Index: 1, URL: "https://youtu.be/b"}}
}

func (r *twoLinkRun) RemoveSource(playlist int) bool {
	r.removed = append(r.removed, playlist)
	return true
}

func TestALinkCanBeTakenOutWhileTheRunGoesOn(t *testing.T) {
	run := &twoLinkRun{addingRun: &addingRun{fakeRun: newFakeRun()}}
	list := downloadlist.New([]string{"https://youtu.be/a", "https://youtu.be/b"})
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, list))
	m = m.runStartedFor(t, run)

	next, _ := m.Update(downloadpage.ListEditMsg{Op: "remove", Index: 1})
	m = next.(Model)
	if len(run.removed) != 1 || run.removed[0] != 1 {
		t.Errorf("the run was told to remove %v", run.removed)
	}
	if got := addressesOf(m); len(got) != 1 || got[0] != "https://youtu.be/a" {
		t.Errorf("the list has %v", got)
	}
	if m.starting || !m.running() {
		t.Error("taking a link out must not restart the run")
	}
}

// busyRun is a fake run with songs that are downloading.
type busyRun struct{ *addingRun }

func (busyRun) Sources() []engine.PlaylistSource {
	return []engine.PlaylistSource{{Index: 0, URL: "https://youtu.be/a"}}
}

func TestClosingAsksOnlyWhileSongsAreDownloading(t *testing.T) {
	run := busyRun{&addingRun{fakeRun: newFakeRun()}}
	list := downloadlist.New([]string{"https://youtu.be/a"})
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, list))
	m = m.runStartedFor(t, run)
	m.screen.State().Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1}}})

	// Songs wait to be chosen: nothing downloads, so nothing is asked.
	m.screen.State().Apply(engine.EvEntryState{Entry: 0, State: engine.StatePlanned, TargetID: "abc12345678"})
	next, cmd := m.Update(CloseRequestedMsg{})
	if next.(Model).closing.On || cmd == nil {
		t.Fatalf("a close with nothing downloading asked %v, cmd %v", next.(Model).closing.On, cmd != nil)
	}

	// A song downloads: the app asks, in its own screen.
	m.screen.State().Apply(engine.EvEntryState{Entry: 0, State: engine.StateDownloading, TargetID: "abc12345678"})
	next, cmd = m.Update(CloseRequestedMsg{})
	m = next.(Model)
	if !m.closing.On || cmd != nil {
		t.Fatalf("a close while downloading did not ask")
	}
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "1 song is still downloading") || !strings.Contains(screen, "Keep downloading") {
		t.Errorf("the question is not on the screen:\n%s", screen)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if next.(Model).closing.On {
		t.Error("esc did not keep the downloads")
	}
}

func TestShiftAndTheArrowsTickSongsEvenThoughTheMenuHasThemForMovingLinks(t *testing.T) {
	run := busyRun{&addingRun{fakeRun: newFakeRun()}}
	list := downloadlist.New([]string{"https://youtu.be/a"})
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, list))
	m = m.runStartedFor(t, run)
	entries := []engine.EntryInfo{{ID: 0, Playlist: 0, Index: 1, Title: "one"}, {ID: 1, Playlist: 0, Index: 2, Title: "two"}, {ID: 2, Playlist: 0, Index: 3, Title: "three"}}
	m.screen.State().Apply(engine.EvPlaylistListed{Playlist: 0, Title: "A", Entries: entries})
	for i := range entries {
		m.screen.State().Apply(engine.EvEntryState{Entry: i, State: engine.StatePlanned, TargetID: "abc12345678"})
	}
	press := func(m Model, key tea.KeyPressMsg) Model {
		next, _ := m.Update(key)

		return next.(Model)
	}
	m = press(m, tea.KeyPressMsg{Code: tea.KeyTab})
	m = press(m, tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift})
	m = press(m, tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift})
	if got := strings.Count(ansi.Strip(m.render()), "[x]"); got != 4 { // the three songs and the box of the heading
		t.Errorf("%d boxes are ticked, want 4:\n%s", got, ansi.Strip(m.render()))
	}
}
