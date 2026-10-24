package download

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Chooser is what a run offers when it holds each song for the person to choose: a song waits,
// found and proposed, until it is accepted, with the video it proposes or another, or rejected. A
// song that was rejected can be accepted again. The engine does when it is told to hold the songs.
type Chooser interface {
	// Release accepts song entry: it goes to download as proposed when videoID is "", else with that
	// video.
	Release(entry int, videoID string) bool
	// ReleaseAll accepts every song that waits, as proposed, and says how many.
	ReleaseAll() int
	// Skip rejects song entry: it is not downloaded.
	Skip(entry int) bool
	// Swap gives song entry another video to download, also when it was saved already: what it
	// saved is then taken away. False when the song is on its way and cannot be swapped.
	Swap(entry int, videoID string) bool
}

// SourceRemover is what a run offers when it takes a link out while it goes on.
type SourceRemover interface {
	// RemoveSource takes playlist (the engine's number for it) out of the run.
	RemoveSource(playlist int) bool
}

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
	if len(m.picked) == 1 && m.picked[id] {
		m.picked = nil

		return
	}
	m.picked = map[int]bool{id: true}
}

// rejectable says whether a song can be rejected now: it waits, or it was accepted and has not begun.
func rejectable(en *runstate.Entry) bool {
	return en != nil && (en.State == engine.StatePlanned || en.State == engine.StateAccepted)
}

// accepted is what the screen says after songs were accepted: they wait for the Download button, or
// they go on at once.
func (m Model) accepted(n int) string {
	if m.paused {
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
	if m.ctrl == nil {
		return nil, false
	}
	c, ok := m.ctrl.(Chooser)

	return c, ok
}

// anyWaiting says whether some song waits for the person to choose it.
func (m Model) anyWaiting() bool {
	_, ok := m.chooser()

	return ok && m.state.Tally().Planned > 0
}

// targetIDs are the songs the actions work on: the ones ticked, else the one the cursor is on.
func (m Model) targetIDs() []int {
	if len(m.picked) > 0 {
		var ids []int
		for id := range m.picked {
			if m.state.Entry(id) != nil {
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
	picked := make(map[int]bool, len(m.picked)+1)
	for k := range m.picked {
		picked[k] = true
	}
	if picked[id] {
		delete(picked, id)
	} else {
		picked[id] = true
	}
	m.picked = picked
}

// togglePickAll ticks every song of the table, or unticks them all when they all are.
func (m *Model) togglePickAll() {
	ids := m.visibleEntries()
	all := len(ids) > 0
	for _, id := range ids {
		if !m.picked[id] {
			all = false
		}
	}
	picked := make(map[int]bool, len(m.picked))
	for k := range m.picked {
		picked[k] = true
	}
	for _, id := range ids {
		if all {
			delete(picked, id)
		} else {
			picked[id] = true
		}
	}
	m.picked = picked
}

// pickState is how ticked the songs of the table are: 0 none, 1 some, 2 all.
func (m Model) pickState() int {
	ids := m.visibleEntries()
	n := 0
	for _, id := range ids {
		if m.picked[id] {
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
		m.picked = nil

		return m, nil, true
	}
	if key == "d" && m.focus == paneLists {
		sent := 0
		if m.selPlaylist < len(m.state.Playlists) {
			for _, id := range m.state.Playlists[m.selPlaylist].Entries {
				if waiting(m.state.Entries[id]) && chooser.Release(id, "") {
					sent++
				}
			}
		}
		m.flash(m.accepted(sent))

		return m, nil, true
	}
	if key == "space" {
		if en := m.selectedEntry(); en != nil && m.focus == paneEntries {
			m.togglePick(en.ID)
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
		en := m.state.Entry(ids[0])
		video := en.TargetID
		if video == "" {
			video = en.VideoID
		}

		return m, func() tea.Msg { return OpenVideoMsg{URL: "https://www.youtube.com/watch?v=" + video} }, true
	case "d":
		done := 0
		for _, id := range ids {
			if acceptable(m.state.Entry(id)) && chooser.Release(id, "") {
				done++
			}
		}
		if done == 0 {
			m.flash("nothing to accept: the song is on its way, or has no video")
		} else {
			m.flash(m.accepted(done))
		}
		m.picked = nil
	case "x":
		done := 0
		for _, id := range ids {
			if rejectable(m.state.Entry(id)) && chooser.Skip(id) {
				done++
			}
		}
		if done == 0 {
			m.flash("nothing to reject: the song is on its way or done")
		} else {
			m.flash(fmt.Sprintf("%s rejected: they will not download. Accept brings them back", counted(done, "song")))
		}
		m.picked = nil
	case "s":
		en := m.state.Entry(ids[0])
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
		m.swapping, m.swap, m.swapEntry = true, in, en.ID

		return m, blink, true
	}

	return m, nil, true
}

// updateSwapping handles a key while the address of another video is being asked: enter sends
// the song to download with that video, esc gives up.
func (m Model) updateSwapping(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.swapping = false

		return m, nil
	case "enter":
		m.swapping = false
		id, ok := playlisteditor.VideoID(strings.TrimSpace(m.swap.Value()))
		if !ok {
			m.flash("that is not the address of a YouTube video")

			return m, nil
		}
		chooser, held := m.chooser()
		switch {
		case !held:
		case chooser.Swap(m.swapEntry, id):
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
