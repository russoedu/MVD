package download

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Chooser is what a run offers when it holds each song for the person to choose: a song waits,
// found and proposed, until it is sent to download, with the video it proposes or another, or
// cancelled. The engine does when it is told to hold the songs.
type Chooser interface {
	// Release sends song entry to download: as proposed when videoID is "", else that video.
	Release(entry int, videoID string) bool
	// ReleaseAll sends every song that waits, as proposed, and says how many.
	ReleaseAll() int
	// Skip cancels song entry: it is not downloaded.
	Skip(entry int) bool
}

// OpenVideoMsg asks the app to show a video in the person's browser.
type OpenVideoMsg struct{ URL string }

// waiting says whether a song waits for the person to choose it.
func waiting(en *runstate.Entry) bool { return en != nil && en.State == engine.StatePlanned }

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

// chooseKey does what a key of the choosing says: d sends the song (or every waiting song of the
// link) to download, D sends all of them, x cancels the song, s swaps its video and o opens it.
// handled is false for any other key, and for a run that does not hold its songs.
func (m Model) chooseKey(key string) (next Model, cmd tea.Cmd, handled bool) {
	if key != "d" && key != "D" && key != "x" && key != "s" && key != "o" {
		return m, nil, false
	}
	chooser, ok := m.chooser()
	if !ok {
		return m, nil, false
	}
	if key == "D" {
		m.flash(fmt.Sprintf("%s sent to download", counted(chooser.ReleaseAll(), "song")))

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
		m.flash(fmt.Sprintf("%s of this link sent to download", counted(sent, "song")))

		return m, nil, true
	}
	en := m.selectedEntry()
	if en == nil {
		m.flash("choose a song first")

		return m, nil, true
	}
	switch key {
	case "o":
		video := en.TargetID
		if video == "" {
			video = en.VideoID
		}

		return m, func() tea.Msg { return OpenVideoMsg{URL: "https://www.youtube.com/watch?v=" + video} }, true
	case "d":
		switch {
		case !waiting(en):
			m.flash("this song is not waiting for you")
		case chooser.Release(en.ID, ""):
			m.flash("sent to download")
		default:
			m.flash("nothing was found for this song: swap its video or cancel it")
		}
	case "x":
		if chooser.Skip(en.ID) {
			m.flash("cancelled: it will not be downloaded")
		} else {
			m.flash("this song cannot be cancelled now")
		}
	case "s":
		if !waiting(en) && en.State != engine.StateFailed && en.State != engine.StateNotFound {
			m.flash("this song is already on its way")

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
		case chooser.Release(m.swapEntry, id):
			m.flash("sent to download with the video you chose")
		default:
			m.flash("this song cannot be changed now")
		}

		return m, nil
	}
	var cmd tea.Cmd
	m.swap, cmd = m.swap.Update(msg)

	return m, cmd
}
