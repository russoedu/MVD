package workspace

import (
	"testing"

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

func TestALinkAddedWhileLoadingIsWrittenInTheSameRowAndGoesToTheRun(t *testing.T) {
	run := &addingRun{fakeRun: newFakeRun()}
	m := sized(New(autoLoadHost(t, run), config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m = m.runStartedFor(t, run)

	m, _ = choose(m, entryAddLink)
	if !m.inline.on || !m.inline.running {
		t.Fatal("the row for the new link did not open")
	}
	for _, r := range "https://youtu.be/second" {
		next, _ := m.Update(tea.KeyPressMsg{Text: string(r), Code: r})
		m = next.(Model)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(Model)
	if m.inline.on || len(run.added) != 1 || run.added[0] != "https://youtu.be/second" {
		t.Errorf("inline %v, added %v", m.inline.on, run.added)
	}
	if got := len(m.screen.State().Playlists); got != 0 {
		t.Errorf("the row that was made up for writing is still there: %d playlists", got)
	}
	if got := addressesOf(m); len(got) != 2 {
		t.Errorf("the list has %v", got)
	}
}

// runStartedFor gives the model the run its start was waiting for.
func (m Model) runStartedFor(t *testing.T, run ListRun) Model {
	t.Helper()
	next, _ := m.Update(runStartedMsg{run: run})

	return next.(Model)
}
