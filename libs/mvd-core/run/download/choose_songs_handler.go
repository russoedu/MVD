package download

import (
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Chooser is what a run offers when it holds each song for the person to choose.
type Chooser = downloadsession.Chooser

// SourceRemover is what a run offers when it takes a link out while it goes on.
type SourceRemover = downloadsession.SourceRemover

// OpenVideoMsg asks the app to show a video in the person's browser.
type OpenVideoMsg struct{ URL string }

// waiting says whether a song waits for the person to choose it.
func waiting(en *runstate.Entry) bool { return en != nil && en.State == engine.StatePlanned }

// swappable says whether the video of a song can be swapped now: it waits, was saved, failed or was
// rejected, and is not on its way.
func swappable(en *runstate.Entry) bool {
	if en == nil {
		return false
	}
	switch en.State {
	case engine.StatePlanned, engine.StateDone, engine.StateDuplicate, engine.StateFailed, engine.StateNotFound, engine.StateSkipped:
		return true
	}

	return false
}

// pickOnly is a click on a song, away from its box: it alone is ticked, or when it was the only one
// ticked the click takes the tick away.
func (m *Model) pickOnly(id int) {
	if len(m.session.Picked) == 1 && m.session.Picked[id] {
		m.session.Picked = nil

		return
	}
	m.session.Picked = map[int]bool{id: true}
}

// rejectable says whether a song can be rejected now: it waits, or it was accepted and has not begun.
func rejectable(en *runstate.Entry) bool {
	return en != nil && (en.State == engine.StatePlanned || en.State == engine.StateAccepted)
}

// accepted is what the screen says after songs were accepted: they wait for the Download button, or
// they go on at once.
func (m Model) accepted(n int) string {
	if m.session.Paused {
		return counted(n, "song") + " accepted. Press Download to start them"
	}

	return counted(n, "song") + " accepted: they will download"
}

// acceptable says whether a song can be accepted now: it waits, or it was rejected.
func acceptable(en *runstate.Entry) bool {
	return en != nil && (en.State == engine.StatePlanned || (en.State == engine.StateSkipped && en.TargetID != ""))
}

// chooser is the run as something that holds songs for choosing, when it is.
func (m Model) chooser() (Chooser, bool) {
	if m.session.Ctrl == nil {
		return nil, false
	}
	c, ok := m.session.Ctrl.(Chooser)

	return c, ok
}

// anyWaiting says whether some song waits for the person to choose it.
func (m Model) anyWaiting() bool {
	_, ok := m.chooser()

	return ok && m.session.State.Tally().Planned > 0
}

// targetIDs are the songs the actions work on: the ones ticked, else the one the cursor is on.
func (m Model) targetIDs() []int {
	if len(m.session.Picked) > 0 {
		var ids []int
		for id := range m.session.Picked {
			if m.session.State.Entry(id) != nil {
				ids = append(ids, id)
			}
		}
		sort.Ints(ids)

		return ids
	}
	if en := m.selectedEntry(); en != nil {
		return []int{en.ID}
	}

	return nil
}

// togglePick ticks or unticks a song.
func (m *Model) togglePick(id int) {
	picked := make(map[int]bool, len(m.session.Picked)+1)
	for k := range m.session.Picked {
		picked[k] = true
	}
	if picked[id] {
		delete(picked, id)
	} else {
		picked[id] = true
	}
	m.session.Picked = picked
}

// togglePickAll ticks every song of the table, or unticks them all when they all are.
func (m *Model) togglePickAll() {
	ids := m.visibleEntries()
	all := len(ids) > 0
	for _, id := range ids {
		if !m.session.Picked[id] {
			all = false
		}
	}
	picked := make(map[int]bool, len(m.session.Picked))
	for k := range m.session.Picked {
		picked[k] = true
	}
	for _, id := range ids {
		if all {
			delete(picked, id)
		} else {
			picked[id] = true
		}
	}
	m.session.Picked = picked
}

// pickRange ticks, or unticks, the songs from one song to another, in the order they are shown.
func (m *Model) pickRange(from, to int, on bool) {
	ids := m.visibleEntries()
	a, b := -1, -1
	for i, id := range ids {
		if id == from {
			a = i
		}
		if id == to {
			b = i
		}
	}
	if a < 0 || b < 0 {
		return
	}
	if a > b {
		a, b = b, a
	}
	picked := make(map[int]bool, len(m.session.Picked)+b-a+1)
	for k := range m.session.Picked {
		picked[k] = true
	}
	for _, id := range ids[a : b+1] {
		if on {
			picked[id] = true
		} else {
			delete(picked, id)
		}
	}
	m.session.Picked = picked
}

// tickSong is a click on the box of a song: it ticks or unticks that song, and with Shift all the songs
// from the one ticked last, as a list of mail does.
func (m *Model) tickSong(id int, shift bool) {
	on := !m.session.Picked[id]
	if shift && m.session.AnchorSet {
		m.pickRange(m.session.Anchor, id, on)
	} else {
		m.togglePick(id)
	}
	m.session.Anchor, m.session.AnchorSet = id, true
}

// chooseSong is a click on a song away from its box: it alone is ticked, and with Shift every song
// from the one ticked last to it.
func (m *Model) chooseSong(id int, shift bool) {
	if shift && m.session.AnchorSet {
		m.pickRange(m.session.Anchor, id, true)

		return
	}
	m.pickOnly(id)
	m.session.Anchor, m.session.AnchorSet = id, true
}

// extendPick moves the cursor by delta songs with Shift held, and ticks every song from where the
// range began to the one the cursor is on.
func (m *Model) extendPick(delta int) {
	en := m.selectedEntry()
	if en == nil {
		return
	}
	if !m.session.AnchorSet {
		m.session.Anchor, m.session.AnchorSet = en.ID, true
	}
	m.move(delta)
	if now := m.selectedEntry(); now != nil {
		m.session.Picked = nil
		m.pickRange(m.session.Anchor, now.ID, true)
	}
}

// pickState is how ticked the songs of the table are: 0 none, 1 some, 2 all.
func (m Model) pickState() int {
	ids := m.visibleEntries()
	n := 0
	for _, id := range ids {
		if m.session.Picked[id] {
			n++
		}
	}
	switch {
	case n == 0:
		return 0
	case n == len(ids):
		return 2
	}

	return 1
}

// chooseKey does what a key of the choosing says: d accepts the songs, x rejects them, o opens the
// video of one, space ticks the song the cursor is on and D accepts every song that waits. On a link
// d accepts all its waiting songs. handled is false for any other key, and for a run that does not
// hold its songs.
func (m Model) chooseKey(key string) (next Model, cmd tea.Cmd, handled bool) {
	if key != "d" && key != "D" && key != "x" && key != "s" && key != "o" && key != "space" {
		return m, nil, false
	}
	chooser, ok := m.chooser()
	if !ok {
		return m, nil, false
	}
	if key == "D" {
		m.flash(m.accepted(chooser.ReleaseAll()))
		m.session.Picked = nil

		return m, nil, true
	}
	if key == "d" && m.session.Pane == paneLists && len(m.session.Picked) == 0 {
		sent := 0
		if m.session.SelPlaylist < len(m.session.State.Playlists) {
			for _, id := range m.session.State.Playlists[m.session.SelPlaylist].Entries {
				if waiting(m.session.State.Entries[id]) && chooser.Release(id, "") {
					sent++
				}
			}
		}
		m.flash(m.accepted(sent))

		return m, nil, true
	}
	if key == "space" {
		if en := m.selectedEntry(); en != nil && m.session.Pane == paneEntries {
			m.tickSong(en.ID, false)
		}

		return m, nil, true
	}
	ids := m.targetIDs()
	if len(ids) == 0 {
		m.flash("choose a song first")

		return m, nil, true
	}
	switch key {
	case "o":
		if len(ids) != 1 {
			m.flash("open one video at a time")

			return m, nil, true
		}
		en := m.session.State.Entry(ids[0])
		video := en.TargetID
		if video == "" {
			video = en.VideoID
		}

		return m, func() tea.Msg { return OpenVideoMsg{URL: "https://www.youtube.com/watch?v=" + video} }, true
	case "d":
		done := 0
		for _, id := range ids {
			if acceptable(m.session.State.Entry(id)) && chooser.Release(id, "") {
				done++
			}
		}
		if done == 0 {
			m.flash("nothing to accept: the song is on its way, or has no video")
		} else {
			m.flash(m.accepted(done))
		}
		m.session.Picked = nil
	case "x":
		done := 0
		for _, id := range ids {
			if rejectable(m.session.State.Entry(id)) && chooser.Skip(id) {
				done++
			}
		}
		if done == 0 {
			m.flash("nothing to reject: the song is on its way or done")
		} else {
			m.flash(fmt.Sprintf("%s rejected: they will not download. Accept brings them back", counted(done, "song")))
		}
		m.session.Picked = nil
	case "s":
		en := m.session.State.Entry(ids[0])
		switch {
		case len(ids) != 1:
			m.flash("swap the video of one song at a time")

			return m, nil, true
		case !swappable(en):
			m.flash("this song is on its way: swap it when it is saved")

			return m, nil, true
		}
		in := textinput.New()
		in.Prompt = "Address of the video to download instead: "
		in.CharLimit = 0
		in.SetWidth(max(m.width-len(in.Prompt)-6, 10))
		blink := in.Focus()
		m.session.Swapping, m.swap, m.session.SwapEntry = true, in, en.ID

		return m, blink, true
	}

	return m, nil, true
}

// updateSwapping handles a key while the address of another video is being asked: enter sends
// the song to download with that video, esc gives up.
func (m Model) updateSwapping(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.session.Swapping = false

		return m, nil
	case "enter":
		m.session.Swapping = false
		id, ok := playlisteditor.VideoID(strings.TrimSpace(m.swap.Value()))
		if !ok {
			m.flash("that is not the address of a YouTube video")

			return m, nil
		}
		chooser, held := m.chooser()
		switch {
		case !held:
		case chooser.Swap(m.session.SwapEntry, id):
			m.flash("accepted with the video you chose")
		default:
			m.flash("this song cannot be changed now")
		}

		return m, nil
	}
	var cmd tea.Cmd
	m.swap, cmd = m.swap.Update(msg)

	return m, cmd
}
