package download

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/testing/keys"
)

// choosingRun is a fake run that holds its songs and remembers what was done with them.
type choosingRun struct {
	*fakeController
	released map[int]string
	skipped  []int
	all      int
}

func (c *choosingRun) Release(entry int, video string) bool {
	if c.released == nil {
		c.released = map[int]string{}
	}
	c.released[entry] = video

	return true
}
func (c *choosingRun) ReleaseAll() int     { c.all++; return 3 }
func (c *choosingRun) Skip(entry int) bool { c.skipped = append(c.skipped, entry); return true }

// choosingModel has the fourth song of the sample waiting for the person.
func choosingModel(t *testing.T) (Model, *choosingRun) {
	t.Helper()
	m := sampleModel(t)
	run := &choosingRun{fakeController: m.ctrl.(*fakeController)}
	m.ctrl = run
	m.state.Apply(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "abc12345678", Official: true})
	m.focus, m.selPlaylist, m.selEntry = paneEntries, 0, 3

	return m, run
}

func pressKey(m Model, key string) Model {
	next, _ := m.Update(keys.Press(key))

	return next.(Model)
}

func TestDSendsTheSelectedSongToDownloadAndXCancelsIt(t *testing.T) {
	m, run := choosingModel(t)
	m = pressKey(m, "d")
	if _, ok := run.released[3]; !ok || run.released[3] != "" {
		t.Errorf("released %v", run.released)
	}
	m = pressKey(m, "x")
	if len(run.skipped) != 1 || run.skipped[0] != 3 {
		t.Errorf("skipped %v", run.skipped)
	}
	m = pressKey(m, "D")
	if run.all != 1 {
		t.Errorf("release all was called %d times", run.all)
	}
	if !strings.Contains(m.status, "3 songs") {
		t.Errorf("the screen says %q", m.status)
	}
}

func TestSAsksForAnotherVideoAndSendsTheSongWithIt(t *testing.T) {
	m, run := choosingModel(t)
	m = pressKey(m, "s")
	if !m.swapping {
		t.Fatal("the address of another video was not asked")
	}
	for _, r := range "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		m = pressKey(m, string(r))
	}
	m = pressKey(m, "enter")
	if m.swapping || run.released[3] != "dQw4w9WgXcQ" {
		t.Errorf("swapping %v, released %v", m.swapping, run.released)
	}
	m = pressKey(m, "s")
	m = pressKey(m, "x") // typed into the address, not a cancel
	m = pressKey(m, "esc")
	if m.swapping || len(run.skipped) != 0 {
		t.Errorf("swapping %v, skipped %v", m.swapping, run.skipped)
	}
}

func TestTheWaitingSongsAreSaidInWordsAndTheKeysAreOffered(t *testing.T) {
	m, _ := choosingModel(t)
	if status, _ := entryStatus(m.state.Entries[3]); status != "to choose" {
		t.Errorf("status %q", status)
	}
	counters, _ := m.headerTexts()
	if !strings.Contains(counters, "To choose 1") {
		t.Errorf("header %q", counters)
	}
	var keysOffered []string
	for _, h := range m.hints() {
		keysOffered = append(keysOffered, h.Key)
	}
	if got := strings.Join(keysOffered, " "); !strings.HasPrefix(got, "d s x o D") {
		t.Errorf("keys %q", got)
	}
	sentence, _ := m.songSentence(m.state.Entries[3])
	if !strings.Contains(sentence, "official music video") || !strings.Contains(sentence, "Choose what to do") {
		t.Errorf("sentence %q", sentence)
	}
}

func TestTheChoicesOfAWaitingSongAreButtonsThatCanBeClicked(t *testing.T) {
	m, run := choosingModel(t)
	m.width, m.height = 120, 30
	buttons := m.choiceButtons()
	if len(buttons) != 4 {
		t.Fatalf("buttons %+v", buttons)
	}
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "Swap video") || !strings.Contains(screen, "Cancel song") {
		t.Errorf("the buttons are not drawn:\n%s", screen)
	}
	click := func(m Model, b choiceButton) Model {
		next, _ := m.Update(tea.MouseClickMsg{X: b.x + 1, Y: b.y, Button: tea.MouseLeft})

		return next.(Model)
	}
	m = click(m, buttons[2]) // Cancel song
	if len(run.skipped) != 1 {
		t.Errorf("skipped %v", run.skipped)
	}
	m = click(m, buttons[1]) // Swap video
	if !m.swapping {
		t.Error("the swap did not ask for an address")
	}
}
