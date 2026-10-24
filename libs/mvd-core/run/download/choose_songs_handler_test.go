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
func (c *choosingRun) Swap(entry int, video string) bool { return c.Release(entry, video) }
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

func TestSwappingTheVideoIsNotAvailableYet(t *testing.T) {
	m, run := choosingModel(t)
	m = pressKey(m, "s")
	if m.swapping || len(run.released) != 0 || !strings.Contains(m.status, "not available") {
		t.Errorf("swapping %v, released %v, status %q", m.swapping, run.released, m.status)
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
	if got := strings.Join(keysOffered, " "); !strings.HasPrefix(got, "space d x o D") {
		t.Errorf("keys %q", got)
	}
	sentence, _ := m.songSentence(m.state.Entries[3])
	if !strings.Contains(sentence, "official music video") || !strings.Contains(sentence, "Accept it") {
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
	if screen := ansi.Strip(m.render()); !strings.Contains(screen, "Swap video") || !strings.Contains(screen, "Reject") {
		t.Errorf("the buttons are not drawn:\n%s", screen)
	}
	click := func(m Model, b choiceButton) Model {
		next, _ := m.Update(tea.MouseClickMsg{X: b.x + 1, Y: b.y, Button: tea.MouseLeft})

		return next.(Model)
	}
	m = click(m, buttons[1]) // Reject
	if len(run.skipped) != 1 {
		t.Errorf("skipped %v", run.skipped)
	}
	m = click(m, buttons[2]) // Swap video, which cannot be used yet
	if m.swapping || !buttons[2].disabled {
		t.Error("the swap button must be disabled")
	}
}

func TestSongsTickedWithTheBoxesAreAcceptedOrRejectedTogether(t *testing.T) {
	m, run := choosingModel(t)
	m.state.Apply(engine.EvEntryState{Entry: 1, State: engine.StatePlanned, TargetID: "bbb12345678"})
	m.state.Apply(engine.EvEntryState{Entry: 2, State: engine.StatePlanned, TargetID: "ccc12345678"})
	m.width, m.height = 120, 30
	m.selEntry = 1
	m = pressKey(m, "space")
	m.selEntry = 2
	m = pressKey(m, "space")
	if len(m.picked) != 2 || m.pickState() != 1 {
		t.Fatalf("picked %v, state %d", m.picked, m.pickState())
	}
	buttons := m.choiceButtons()
	if buttons[0].label != "✓ Accept 2" || !buttons[3].disabled {
		t.Errorf("buttons %+v", buttons)
	}
	m = pressKey(m, "o")
	if !strings.Contains(m.status, "one video at a time") {
		t.Errorf("opening many says %q", m.status)
	}
	m = pressKey(m, "x")
	if len(run.skipped) != 2 || len(m.picked) != 0 {
		t.Errorf("skipped %v, picked %v", run.skipped, m.picked)
	}
	m.togglePickAll()
	if m.pickState() != 2 {
		t.Errorf("select all gives state %d", m.pickState())
	}
	m.togglePickAll()
	if m.pickState() != 0 {
		t.Errorf("select all again gives state %d", m.pickState())
	}
}

func TestARejectedSongCanBeAcceptedAgain(t *testing.T) {
	m, run := choosingModel(t)
	m.state.Apply(engine.EvEntryState{Entry: 3, State: engine.StateSkipped, TargetID: "abc12345678"})
	if !acceptable(m.state.Entry(3)) {
		t.Fatal("a rejected song is not offered to be accepted")
	}
	m = pressKey(m, "d")
	if _, ok := run.released[3]; !ok {
		t.Errorf("released %v", run.released)
	}
}

func TestACarriedBorderResizesTheColumnAndTheListOfColumnsHidesOne(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	l, _ := m.layout()
	ix, iy, iw := m.tableInner(l)
	spans := m.columnSpans(iw)
	var artist columnSpan
	for _, s := range spans {
		if s.name == columnArtist {
			artist = s
		}
	}
	before := artist.x1 - artist.x0
	press := tea.MouseClickMsg{X: ix + artist.x1, Y: iy, Button: tea.MouseLeft}
	next, _ := m.Update(press)
	m = next.(Model)
	if !m.colDrag.on {
		t.Fatal("the border was not taken")
	}
	next, _ = m.Update(tea.MouseMotionMsg{X: ix + artist.x1 + 5, Y: iy, Button: tea.MouseLeft})
	m = next.(Model)
	next, _ = m.Update(tea.MouseReleaseMsg{X: ix + artist.x1 + 5, Y: iy, Button: tea.MouseLeft})
	m = next.(Model)
	if got := m.colWidth[columnArtist]; got != before+5 || m.colDrag.on {
		t.Errorf("artist is %d wide (was %d), dragging %v", got, before, m.colDrag.on)
	}

	m = pressKey(m, "c")
	if !m.columnList.open {
		t.Fatal("the list of columns did not open")
	}
	m = pressKey(m, "space") // the first one: Artist
	m = pressKey(m, "esc")
	if names := namesOf(m.tableColumns(140)); strings.Contains(names, columnArtist) {
		t.Errorf("the artist column is still there: %s", names)
	}
}
