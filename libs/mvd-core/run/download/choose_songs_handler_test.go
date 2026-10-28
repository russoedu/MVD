package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"sort"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/songlist"
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
func (c *choosingRun) ReleaseAll() int                   { c.all++; return 3 }
func (c *choosingRun) Swap(entry int, video string) bool { return c.Release(entry, video) }
func (c *choosingRun) Skip(entry int) bool               { c.skipped = append(c.skipped, entry); return true }

// choosingModel has the fourth song of the sample waiting for the person.
func choosingModel(t *testing.T) (Model, *choosingRun) {
	t.Helper()
	m := sampleModel(t)
	run := &choosingRun{fakeController: m.session.Ctrl.(*fakeController)}
	m.session.Ctrl = run
	m.session.State.Apply(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "abc12345678", Official: true})
	m.session.Pane, m.session.SelPlaylist, m.session.SelEntry = paneEntries, 0, 3

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
	if !strings.Contains(m.session.Notice, "3 songs") {
		t.Errorf("the screen says %q", m.session.Notice)
	}
}

func TestSwapAsksForAnotherVideoAndSendsTheSongWithIt(t *testing.T) {
	m, run := choosingModel(t)
	m = pressKey(m, "s")
	if !m.session.Swapping {
		t.Fatal("the address of another video was not asked")
	}
	for _, r := range "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		m = pressKey(m, string(r))
	}
	m = pressKey(m, "enter")
	if m.session.Swapping || run.released[3] != "dQw4w9WgXcQ" {
		t.Errorf("swapping %v, released %v", m.session.Swapping, run.released)
	}
	m.session.State.Apply(engine.EvEntryState{Entry: 3, State: engine.StateDownloading, TargetID: "x"})
	m = pressKey(m, "s")
	if m.session.Swapping {
		t.Error("a song on its way was offered a swap")
	}
}

func TestClickingASongAwayFromItsBoxChoosesThatSongAlone(t *testing.T) {
	m, _ := choosingModel(t)
	m.session.State.Apply(engine.EvEntryState{Entry: 1, State: engine.StatePlanned, TargetID: "bbb12345678"})
	m.width, m.height = 140, 30
	l, _ := m.layout()
	ix, iy, iw := m.tableInner(l)
	var box, song columnSpan
	for _, s := range m.columnSpans(iw) {
		switch s.name {
		case columnCheck:
			box = s
		case columnSong:
			song = s
		}
	}
	rowY := func(entry int) int { return iy + 1 + entry }
	click := func(m Model, x, y int) Model {
		next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})

		return next.(Model)
	}
	m = click(m, ix+box.x0+1, rowY(0)) // the box of the first song
	m = click(m, ix+box.x0+1, rowY(1)) // the box of the second
	if len(m.session.Picked) != 2 {
		t.Fatalf("picked %v", m.session.Picked)
	}
	m = click(m, ix+song.x0+2, rowY(1)) // the second song itself: it alone
	if len(m.session.Picked) != 1 || !m.session.Picked[1] {
		t.Errorf("picked %v", m.session.Picked)
	}
	m = click(m, ix+song.x0+2, rowY(1)) // again: the tick goes
	if len(m.session.Picked) != 0 {
		t.Errorf("picked %v", m.session.Picked)
	}
}

func TestTheWaitingSongsAreSaidInWordsAndTheKeysAreOffered(t *testing.T) {
	m, _ := choosingModel(t)
	if status, _ := entryStatus(m.session.State.Entries[3]); status != "official found" {
		t.Errorf("status %q", status)
	}
	counters, _ := m.headerTexts()
	if !strings.Contains(counters, "Official found 1") {
		t.Errorf("header %q", counters)
	}
	var keysOffered []string
	for _, h := range m.hints() {
		keysOffered = append(keysOffered, h.Key)
	}
	if got := strings.Join(keysOffered, " "); !strings.HasPrefix(got, "space d x o D") {
		t.Errorf("keys %q", got)
	}
	sentence, _ := m.songSentence(m.session.State.Entries[3])
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
	if buttons[2].disabled {
		t.Error("the swap button must be usable with one song")
	}
	m = click(m, buttons[2]) // Swap video
	if !m.session.Swapping {
		t.Error("the swap did not ask for an address")
	}
}

func TestSongsTickedWithTheBoxesAreAcceptedOrRejectedTogether(t *testing.T) {
	m, run := choosingModel(t)
	m.session.State.Apply(engine.EvEntryState{Entry: 1, State: engine.StatePlanned, TargetID: "bbb12345678"})
	m.session.State.Apply(engine.EvEntryState{Entry: 2, State: engine.StatePlanned, TargetID: "ccc12345678"})
	m.width, m.height = 120, 30
	m.session.SelEntry = 1
	m = pressKey(m, "space")
	m.session.SelEntry = 2
	m = pressKey(m, "space")
	if len(m.session.Picked) != 2 || m.session.PickState() != 1 {
		t.Fatalf("picked %v, state %d", m.session.Picked, m.session.PickState())
	}
	buttons := m.choiceButtons()
	if buttons[0].label != "✓ Accept 2" || !buttons[3].disabled {
		t.Errorf("buttons %+v", buttons)
	}
	m = pressKey(m, "o")
	if !strings.Contains(m.session.Notice, "one video at a time") {
		t.Errorf("opening many says %q", m.session.Notice)
	}
	m = pressKey(m, "x")
	if len(run.skipped) != 2 || len(m.session.Picked) != 0 {
		t.Errorf("skipped %v, picked %v", run.skipped, m.session.Picked)
	}
	m.session.TickAll()
	if m.session.PickState() != 2 {
		t.Errorf("select all gives state %d", m.session.PickState())
	}
	m.session.TickAll()
	if m.session.PickState() != 0 {
		t.Errorf("select all again gives state %d", m.session.PickState())
	}
}

func TestARejectedSongCanBeAcceptedAgain(t *testing.T) {
	m, run := choosingModel(t)
	m.session.State.Apply(engine.EvEntryState{Entry: 3, State: engine.StateSkipped, TargetID: "abc12345678"})
	if !downloadsession.Acceptable(m.session.State.Entry(3)) {
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

func TestColumnsCanBeCarriedToAnotherPlaceInTheList(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	m = pressKey(m, "c")
	m = pressKey(m, "shift+down") // Artist goes below Video name
	if got := m.session.ArrangedColumns(); got[0] != columnVideo || got[1] != columnArtist {
		t.Errorf("order %v", got)
	}
	if names := namesOf(m.tableColumns(140)); strings.Index(names, columnArtist) < strings.Index(names, columnVideo) && strings.Contains(names, columnVideo) {
		t.Errorf("the table does not follow the order: %s", names)
	}

	// With the mouse: the grip of the first row carried to the third.
	m.columnList = columnsMenu{open: true}
	l, _ := m.layout()
	r := m.columnsMenuRect(l)
	for _, msg := range []tea.MouseMsg{
		tea.MouseClickMsg{X: r.x + 2, Y: r.y + 1, Button: tea.MouseLeft},
		tea.MouseMotionMsg{X: r.x + 2, Y: r.y + 3, Button: tea.MouseLeft},
		tea.MouseReleaseMsg{X: r.x + 2, Y: r.y + 3, Button: tea.MouseLeft},
	} {
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	if got := m.session.ArrangedColumns(); got[2] != columnVideo {
		t.Errorf("order after the drag %v", got)
	}
}

func TestAClickOnAHeadingSortsAndAnotherClickTurnsTheOrderRound(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	l, _ := m.layout()
	ix, iy, iw := m.tableInner(l)
	var song columnSpan
	for _, s := range m.columnSpans(iw) {
		if s.name == columnSong {
			song = s
		}
	}
	click := func(m Model) Model {
		next, _ := m.Update(tea.MouseClickMsg{X: ix + song.x0 + 1, Y: iy, Button: tea.MouseLeft})

		return next.(Model)
	}
	names := func(m Model) []string {
		var out []string
		for _, id := range m.session.VisibleEntries() {
			out = append(out, strings.ToLower(songName(m.session.State.Entries[id])))
		}

		return out
	}
	selected := m.session.SelectedEntry().ID
	m = click(m)
	asc := names(m)
	if m.session.Order.By != columnSong || m.session.Order.Desc || !sort.StringsAreSorted(asc) {
		t.Fatalf("sorted by %q desc %v: %v", m.session.Order.By, m.session.Order.Desc, asc)
	}
	if m.session.SelectedEntry().ID != selected {
		t.Error("the song the cursor was on changed with the sort")
	}
	m = click(m)
	desc := names(m)
	if !m.session.Order.Desc || desc[0] != asc[len(asc)-1] {
		t.Errorf("the second click did not turn the order round: %v", desc)
	}
	if !strings.Contains(ansi.Strip(m.render()), "Song ▼") {
		t.Error("the heading does not show the direction")
	}
}

func TestTheFilterHidesTheStatusesTheListHides(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	all := len(m.session.VisibleEntries())
	m = pressKey(m, "F")
	if !m.filterList.open {
		t.Fatal("the filter did not open")
	}
	for i, g := range songlist.Groups {
		if g.Key == songlist.GroupSaved {
			m.filterList.sel = i
		}
	}
	m = pressKey(m, "space")
	m = pressKey(m, "esc")
	if got := len(m.session.VisibleEntries()); got != all-1 || !m.session.Filter.Active() {
		t.Errorf("%d of %d songs show with the saved ones hidden", got, all)
	}
	if title, _ := m.entryHeader(); !strings.Contains(title, " of ") {
		t.Errorf("the title does not say it is filtered: %q", title)
	}
	m = pressKey(m, "f") // failed only
	for _, id := range m.session.VisibleEntries() {
		if m.session.State.Entries[id].State != engine.StateFailed {
			t.Errorf("song %d is not failed", id)
		}
	}
	m = pressKey(m, "f")
	if m.session.Filter.Active() {
		t.Error("f did not show every song again")
	}
}

func TestACarriedBorderFollowsThePointerAndTheOtherSideGivesTheRoom(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	l, _ := m.layout()
	ix, iy, iw := m.tableInner(l)
	spans := m.columnSpans(iw)
	var artist, progress columnSpan
	for _, s := range spans {
		switch s.name {
		case columnArtist:
			artist = s
		case columnProgress:
			progress = s
		}
	}
	at := func(msg tea.MouseMsg) { next, _ := m.Update(msg); m = next.(Model) }
	at(tea.MouseClickMsg{X: ix + artist.x1, Y: iy, Button: tea.MouseLeft})
	at(tea.MouseMotionMsg{X: ix + artist.x1 + 3, Y: iy, Button: tea.MouseLeft})
	at(tea.MouseReleaseMsg{X: ix + artist.x1 + 3, Y: iy, Button: tea.MouseLeft})
	after := m.columnSpans(iw)
	for _, s := range after {
		switch s.name {
		case columnArtist:
			if s.x1-s.x0 != artist.x1-artist.x0+3 || s.x1 != artist.x1+3 {
				t.Errorf("artist %+v, was %+v: the border did not follow the pointer", s, artist)
			}
		case columnProgress:
			if s.x1 != progress.x1 || s.x1-s.x0 != progress.x1-progress.x0-3 {
				t.Errorf("progress %+v, was %+v", s, progress)
			}
		case columnSong:
			if s.x1 != spans[3].x1 {
				t.Errorf("the song changed: %+v", s)
			}
		}
	}
}

func TestTheStatusSaysWhetherAnOfficialVideoWasFound(t *testing.T) {
	m, _ := choosingModel(t)
	word := func(en engine.EvEntryState) string {
		m.session.State.Apply(en)
		w, _ := entryStatus(m.session.State.Entries[3])

		return w
	}
	if got := word(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "a", Official: true}); got != "official found" {
		t.Errorf("an official video says %q", got)
	}
	if got := word(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "a", OwnOfficial: true}); got != "official video" {
		t.Errorf("an own official video says %q", got)
	}
	if got := word(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "a", Better: true}); got != "better found" {
		t.Errorf("a better video says %q", got)
	}
	if got := word(engine.EvEntryState{Entry: 3, State: engine.StatePlanned, TargetID: "a"}); got != "original upload" {
		t.Errorf("the original says %q", got)
	}
	if got := word(engine.EvEntryState{Entry: 3, State: engine.StateAccepted, TargetID: "a"}); got != "accepted" {
		t.Errorf("an accepted song says %q", got)
	}
	if got := songlist.GroupOf(m.session.State.Entries[3]); got != songlist.GroupAccepted {
		t.Errorf("an accepted song is in the group %q", got)
	}
}

func TestTheSongsAreSortedByTheNumberFromTheStart(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	if m.sortMark(columnIdx) != "▲" {
		t.Errorf("the # heading says %q", m.sortMark(columnIdx))
	}
	m.sortOn(columnIdx)
	if m.sortMark(columnIdx) != "▼" {
		t.Errorf("a click on the # heading says %q", m.sortMark(columnIdx))
	}
	m.sortOn(columnIdx)
	if m.sortMark(columnIdx) != "▲" {
		t.Errorf("another click says %q", m.sortMark(columnIdx))
	}
}

func TestATickedSongBringsItsButtonsEvenWhileTheLinksHaveTheFocus(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	m.session.Pane = paneLists
	if buttons := m.choiceButtons(); len(buttons) != 1 || !strings.Contains(buttons[0].label, "link") {
		t.Fatalf("with nothing ticked the link's button is there: %+v", buttons)
	}
	m.session.TickSong(3, false)
	buttons := m.choiceButtons()
	if len(buttons) != 4 || buttons[0].label != "✓ Accept" {
		t.Fatalf("a ticked song did not bring its buttons: %+v", buttons)
	}
	if rows := m.detailRows(60); len(rows) == 0 || !strings.Contains(rows[0].Line, "Bob Sinclar") {
		t.Errorf("the details are not about the ticked song: %+v", rows)
	}
	m.session.TickSong(2, false)
	if rows := m.detailRows(60); !strings.Contains(rows[0].Line, "2 songs selected") {
		t.Errorf("many ticked songs are not summed up: %+v", rows[0])
	}
}

func TestALinkThatWasJustAddedIsTheOneSelected(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	next, _ := m.Update(evMsg{ev: engine.EvPlaylistAdded{Source: engine.PlaylistSource{Index: len(m.session.State.Playlists), URL: "https://new"}}})
	m = next.(Model)
	if m.session.SelPlaylist != len(m.session.State.Playlists)-1 || m.session.State.Playlists[m.session.SelPlaylist].URL != "https://new" {
		t.Errorf("selected %d of %d", m.session.SelPlaylist, len(m.session.State.Playlists))
	}

	m.session.SelPlaylist = 1
	attached, _ := m.Attach(m.session.Ctrl)
	if attached.session.SelPlaylist != 1 {
		t.Errorf("a new run moved the cursor from 1 to %d", attached.session.SelPlaylist)
	}
}

func TestShiftTicksTheSongsBetweenTwoClicksAsAListOfMailDoes(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	l, _ := m.layout()
	ix, iy, iw := m.tableInner(l)
	var box columnSpan
	for _, s := range m.columnSpans(iw) {
		if s.name == columnCheck {
			box = s
		}
	}
	click := func(m Model, row int, mod tea.KeyMod) Model {
		next, _ := m.Update(tea.MouseClickMsg{X: ix + box.x0 + 1, Y: iy + 1 + row, Button: tea.MouseLeft, Mod: mod})

		return next.(Model)
	}
	m = click(m, 1, 0)
	m = click(m, 4, tea.ModShift)
	if len(m.session.Picked) != 4 || !m.session.Picked[1] || !m.session.Picked[4] {
		t.Fatalf("shift did not tick the range: %v", m.session.Picked)
	}
	m = click(m, 2, 0) // a click on one box of the range unticks that one alone
	if len(m.session.Picked) != 3 || m.session.Picked[2] {
		t.Errorf("picked %v", m.session.Picked)
	}
	m = click(m, 0, tea.ModShift) // from the box clicked last (2) up to the first: ticks them
	if !m.session.Picked[0] || !m.session.Picked[1] || !m.session.Picked[2] {
		t.Errorf("picked %v", m.session.Picked)
	}
}

func TestShiftAndTheArrowsExtendTheTicks(t *testing.T) {
	m, _ := choosingModel(t)
	m.width, m.height = 140, 30
	m.session.SelEntry = 1
	m = pressKey(m, "shift+down")
	m = pressKey(m, "shift+down")
	if len(m.session.Picked) != 3 || !m.session.Picked[1] || !m.session.Picked[3] {
		t.Fatalf("picked %v", m.session.Picked)
	}
	m = pressKey(m, "shift+up")
	if len(m.session.Picked) != 2 || m.session.Picked[3] {
		t.Errorf("going back did not take the last tick away: %v", m.session.Picked)
	}
}

func TestTheKeyForTheTechnicalLogIsAlwaysOnTheBar(t *testing.T) {
	m, _ := choosingModel(t)
	has := func(m Model, desc string) bool {
		for _, h := range m.hints() {
			if h.Key == "t" && h.Desc == desc {
				return true
			}
		}

		return false
	}
	if !has(m, "technical log") {
		t.Errorf("with a song chosen the bar lacks the technical log: %v", m.hints())
	}
	m.session.Pane = paneLists
	if !has(m, "technical log") {
		t.Errorf("with the links chosen the bar lacks the technical log: %v", m.hints())
	}
	m.session.Technical = true
	if !has(m, "back to the details") {
		t.Errorf("in the log the bar lacks the way back: %v", m.hints())
	}
}
