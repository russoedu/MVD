package download

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/keybar"
)

// choiceButton is one button under the details of the songs that are chosen: what it says, the key it
// stands for, whether it can be used now, and where it is on the screen.
type choiceButton struct {
	label, key string
	disabled   bool
	x, y, w    int
}

// choiceButtonsHeight is the rows the details pane has for itself when buttons are under it.
const choiceButtonsHeight = 9

// choiceButtons are the buttons of what is selected, when the run holds its songs: Accept (the main
// one), Reject, Swap video and Open video for the song or the ticked songs, or for a link the one that
// accepts its songs. They sit on the last row of the details pane. None when there is nothing to
// choose or no room for them.
func (m Model) choiceButtons() []choiceButton {
	if _, held := m.chooser(); !held || m.technical || m.fullLog || m.adding || m.swapping || m.showHelp || m.columnList.open {
		return nil
	}
	l, ok := m.layout()
	if !ok || !l.hasOutput {
		return nil
	}
	var buttons []choiceButton
	switch {
	case m.focus == paneLists && len(m.picked) == 0:
		if m.selPlaylist >= len(m.state.Playlists) || m.waitingIn(m.state.Playlists[m.selPlaylist]) == 0 {
			return nil
		}
		buttons = []choiceButton{{label: "✓ Accept this link's songs", key: "d"}}
	case (m.focus == paneEntries || len(m.picked) > 0) && len(m.targetIDs()) > 0:
		ids := m.targetIDs()
		accept, reject := 0, 0
		for _, id := range ids {
			if acceptable(m.state.Entry(id)) {
				accept++
			}
			if rejectable(m.state.Entry(id)) {
				reject++
			}
		}
		accepting, rejecting := "✓ Accept", "✗ Reject"
		if len(ids) > 1 {
			accepting, rejecting = fmt.Sprintf("✓ Accept %d", accept), fmt.Sprintf("✗ Reject %d", reject)
		}
		buttons = []choiceButton{
			{label: accepting, key: "d", disabled: accept == 0},
			{label: rejecting, key: "x", disabled: reject == 0},
			{label: "⇄ Swap video", key: "s", disabled: len(ids) != 1 || !swappable(m.state.Entry(ids[0]))},
			{label: "↗ Open video", key: "o", disabled: len(ids) != 1},
		}
	default:
		return nil
	}

	inner := l.output.rows()
	x, y := inner.x+1, inner.y+inner.h-1
	var out []choiceButton
	for _, b := range buttons {
		b.w = len([]rune(b.label)) + 4
		if x+b.w > inner.x+inner.w {
			break
		}
		b.x, b.y = x, y
		out = append(out, b)
		x += b.w + 2
	}

	return out
}

// withChoiceButtons draws the buttons over the screen: the first that can be used is the main one.
func (m Model) withChoiceButtons(screen string) string {
	primary := true
	for _, b := range m.choiceButtons() {
		kind := dialog.PillNormal
		switch {
		case b.disabled:
			kind = dialog.PillDisabled
		case primary:
			kind, primary = dialog.PillPrimary, false
		}
		screen = dialog.Overlay(screen, dialog.WidePill(b.label, b.w, kind, false), b.x, b.y)
	}

	return screen
}

// clickChoice does what the button under (x, y) does, when there is one that can be used.
func (m Model) clickChoice(x, y int) (Model, tea.Cmd, bool) {
	for _, b := range m.choiceButtons() {
		if y == b.y && x >= b.x && x < b.x+b.w {
			if b.disabled {
				return m, nil, true
			}
			if key, ok := keybar.KeyMsg(b.key); ok {
				next, cmd := m.handleKey(key)

				return next.(Model), cmd, true
			}
		}
	}

	return m, nil, false
}
