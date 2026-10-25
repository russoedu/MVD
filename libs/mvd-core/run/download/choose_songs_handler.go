package download

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/run/downloadsession"
)

// Chooser is what a run offers when it holds each song for the person to choose.
type Chooser = downloadsession.Chooser

// SourceRemover is what a run offers when it takes a link out while it goes on.
type SourceRemover = downloadsession.SourceRemover

// OpenVideoMsg asks the app to show a video in the person's browser.
type OpenVideoMsg struct{ URL string }

// chooseKey turns a key of the choosing into the intent it stands for: d accepts the songs, x rejects
// them, o opens the video of one, space ticks the song the cursor is on and D accepts every song that
// waits. handled is false for any other key, and for a run that does not hold its songs.
func (m Model) chooseKey(key string) (next Model, cmd tea.Cmd, handled bool) {
	switch key {
	case "d", "D", "x", "s", "o", "space":
	default:
		return m, nil, false
	}
	if !m.session.CanChoose() {
		return m, nil, false
	}
	switch key {
	case "D":
		m.session.AcceptAll()
	case "d":
		m.session.Accept()
	case "x":
		m.session.Reject()
	case "space":
		m.session.TickCursorSong()
	case "o":
		return m, m.run(m.session.OpenVideo()), true
	case "s":
		m.session.BeginSwap()
		if m.session.Swapping {
			in := textinput.New()
			in.Prompt = "Address of the video to download instead: "
			in.CharLimit = 0
			in.SetWidth(max(m.width-len(in.Prompt)-6, 10))
			blink := in.Focus()
			m.swap = in

			return m, blink, true
		}
	}

	return m, nil, true
}

// updateSwapping handles a key while the address of another video is being asked: enter sends
// the song to download with that video, esc gives up.
func (m Model) updateSwapping(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.session.CancelSwap()

		return m, nil
	case "enter":
		m.session.SetSwapText(m.swap.Value())
		m.session.CommitSwap()

		return m, nil
	}
	var cmd tea.Cmd
	m.swap, cmd = m.swap.Update(msg)
	m.session.SetSwapText(m.swap.Value())

	return m, cmd
}
