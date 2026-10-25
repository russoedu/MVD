package workspace

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/menubar"
)

// finishRun makes the run of the model look as one that has ended: its songs are in the state
// the engine would have left them in, and it has said it is idle.
func finishRun(t *testing.T, m Model, done, failed int) Model {
	t.Helper()
	state := m.screen.State()
	state.Apply(engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: len(state.Playlists), URL: "https://a"}})
	state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "a", Entries: entriesOf(done + failed)})
	for i := 0; i < done+failed; i++ {
		s := engine.StateDone
		if i >= done {
			s = engine.StateFailed
		}
		state.Apply(engine.EvEntryState{Entry: i, State: s})
	}
	state.Apply(engine.EvIdle{})

	return m.keepRoomForButtons() // what the next Update of the app does
}

func entriesOf(n int) []engine.EntryInfo {
	out := make([]engine.EntryInfo, n)
	for i := range out {
		out[i] = engine.EntryInfo{ID: i, Playlist: 0, Index: i + 1, VideoID: "v", Title: "song"}
	}

	return out
}

func labels(buttons []button) string {
	var out []string
	for _, b := range buttons {
		out = append(out, b.label)
	}

	return strings.Join(out, "|")
}

func TestARunThatHasNotFinishedIsStillRunningAndNewIsRefused(t *testing.T) {
	m, _ := started(t)
	if m.finished() || !m.active() {
		t.Fatal("a run with work to do is not finished")
	}
	next, _ := m.Update(menubar.SelectedMsg{Name: entryNew})
	if got := next.(Model).notice; !strings.Contains(got, "Stop the downloads") {
		t.Errorf("starting a new list during a run says %q", got)
	}
}

func TestAFinishedRunOffersTheIdleButtonsAndNotPauseOrStop(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 2, 0)
	if !m.finished() || m.active() {
		t.Fatalf("finished %v active %v", m.finished(), m.active())
	}
	got := labels(m.buttons())
	if got != "Open folder|► Load playlists again|+ Add link" {
		t.Errorf("the buttons are %q", got)
	}
	if strings.Contains(got, "Pause") || strings.Contains(got, "Stop") {
		t.Errorf("a finished run still offers %q", got)
	}
}

func TestAFinishedRunWithFailuresOffersRetryToo(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 2)
	if got := labels(m.buttons()); got != "Open folder|► Load playlists again|Retry failed|+ Add link" {
		t.Errorf("the buttons are %q", got)
	}
}

func TestANewListIsAllowedAfterARunFinishedWithoutPressingStop(t *testing.T) {
	m, fake := started(t)
	m = finishRun(t, m, 1, 0)
	next, _ := m.Update(menubar.SelectedMsg{Name: entryNew})
	m = next.(Model)
	if m.running() || !fake.closed {
		t.Errorf("the finished run was not closed: running %v closed %v", m.running(), fake.closed)
	}
	if strings.Contains(m.notice, "Stop the downloads") {
		t.Errorf("a finished run still asks to be stopped: %q", m.notice)
	}
	if len(m.session.Sources()) != 0 {
		t.Errorf("the new list is not empty: %v", m.session.Sources())
	}
}

func TestOpeningAListIsAllowedAfterARunFinished(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 0)
	next, _ := m.Update(menubar.SelectedMsg{Name: entryOpen})
	m = next.(Model)
	if m.running() || strings.Contains(m.notice, "Stop the downloads") {
		t.Errorf("opening a list after a finished run: running %v notice %q", m.running(), m.notice)
	}
	if m.asking != askOpen {
		t.Error("the question for the list to open was not asked")
	}
}

func TestAddingALinkAfterARunFinishedOpensTheRowOnTheList(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 0)
	next, _ := m.Update(menubar.SelectedMsg{Name: entryAddLink})
	m = next.(Model)
	if m.running() || !m.inline.on {
		t.Errorf("running %v, row open %v", m.running(), m.inline.on)
	}
}

func TestStartAgainEndsTheFinishedRunAndBeginsAnother(t *testing.T) {
	m, fake := started(t)
	m = finishRun(t, m, 1, 0)
	next, cmd := m.Update(menubar.SelectedMsg{Name: entryStart})
	m = next.(Model)
	if !fake.closed {
		t.Error("the finished run was not closed")
	}
	if cmd == nil || !m.starting {
		t.Error("another run did not begin to start")
	}
}

func TestAFinishedRunSaysHowItWentAndWhereTheVideosAre(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 3, 1)
	screen := screenOf(m)
	for _, want := range []string{"✓ 3 downloaded", "✗ 1 failed", "Videos in", "Ctrl+Shift+O"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the summary lacks %q:\n%s", want, screen)
		}
	}
}

func TestTheSummaryKeepsTheEndOfALongFolder(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 0)
	m.cfg.OutputDir = `C:\Users\someone\AppData\Local\Temp\a\very\long\path\that\goes\on\and\on\to\MVD-Videos`
	if got := m.summaryLine(); !strings.Contains(ansi.Strip(got), `MVD-Videos`) {
		t.Errorf("the end of the folder is lost: %q", ansi.Strip(got))
	}
}

func TestOpenFolderAsksTheHostAndSaysWhenThereIsNone(t *testing.T) {
	m, _ := started(t)
	dir := t.TempDir()
	m.cfg.OutputDir = dir
	var opened string
	m.host.OpenFolder = func(p string) error { opened = p; return nil }
	next, _ := m.Update(menubar.SelectedMsg{Name: entryOpenFolder})
	if opened != dir {
		t.Errorf("the host was asked to open %q, want %q", opened, dir)
	}

	m = next.(Model)
	m.cfg.OutputDir = dir + "-not-there"
	next, _ = m.Update(menubar.SelectedMsg{Name: entryOpenFolder})
	if got := next.(Model).caution; !strings.Contains(got, "made when the first download starts") {
		t.Errorf("a folder that does not exist yet says %q", got)
	}

	m.host.OpenFolder = nil
	m.cfg.OutputDir = dir
	next, _ = m.Update(menubar.SelectedMsg{Name: entryOpenFolder})
	if got := next.(Model).caution; !strings.Contains(got, "The videos are in") {
		t.Errorf("without a host it says %q", got)
	}
	_ = os.Remove(dir)
}

func TestWhenARunEndsWhatWasSaidAtItsStartMakesRoomForHowItWent(t *testing.T) {
	m, _ := started(t)
	m.done = "Kept this list as a draft: somewhere"
	m = finishRun(t, m, 1, 0)
	next, _ := m.Update(saveTickMsg{})
	m = next.(Model)
	if m.done != "" {
		t.Errorf("the old message is still there: %q", m.done)
	}
	if screen := screenOf(m); !strings.Contains(screen, "1 downloaded") {
		t.Errorf("the summary is not shown: %s", screen)
	}
}

func TestTheSummaryMentionsALinkThatCouldNotBeReadAtAll(t *testing.T) {
	m, _ := started(t)
	m = finishRun(t, m, 1, 0)
	state := m.screen.State()
	state.Apply(engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: len(state.Playlists), URL: "https://bad"}})
	state.Apply(engine.EvPlaylistFailed{Playlist: len(state.Playlists) - 1, Err: "not a valid URL"})
	state.Apply(engine.EvIdle{})
	if got := ansi.Strip(m.summaryLine()); !strings.Contains(got, "1 link could not be read") || !strings.Contains(got, "1 downloaded") {
		t.Errorf("the summary says %q", got)
	}
	if plural("link", 2) != "links" || plural("link", 1) != "link" {
		t.Error("the plural is wrong")
	}
}

func TestAStopSaysHowManySongsWereSavedAndHowManyWereLeft(t *testing.T) {
	m, _ := started(t)
	state := m.screen.State()
	state.Apply(engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: 0, URL: "https://a"}})
	state.Apply(engine.EvPlaylistListed{Playlist: 0, Title: "a", Entries: entriesOf(5)})
	state.Apply(engine.EvEntryState{Entry: 0, State: engine.StateDone})
	state.Apply(engine.EvEntryState{Entry: 1, State: engine.StateDone})

	next, _ := m.Update(runClosedMsg{})
	if got := next.(Model).done; !strings.Contains(got, "Stopped: 2 saved, 3 not downloaded") {
		t.Errorf("a stop says %q", got)
	}
}

func TestStoppingShowsInTheHeaderForAsLongAsTheDownloadsAreHeld(t *testing.T) {
	m, run := started(t)
	m = m.togglePause()
	if !run.paused || !strings.Contains(screenOf(m), "STOPPED") {
		t.Errorf("paused %v, screen:\n%s", run.paused, screenOf(m))
	}
	m = m.togglePause()
	if strings.Contains(screenOf(m), "STOPPED") {
		t.Error("it still says STOPPED after the downloads went on")
	}
}
