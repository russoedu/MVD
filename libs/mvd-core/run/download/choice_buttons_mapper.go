package download

import (
	"charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/keybar"
)

// choiceButton is one button under the details of a song that waits to be chosen: what it says,
// the key it stands for, and where it is on the screen.
type choiceButton struct {
	label, key string
	x, y, w    int
}

// choiceButtonsHeight is the rows the details pane has for itself when buttons are under it.
const choiceButtonsHeight = 9

// choiceButtons are the buttons of what is selected, when a song waits for the person: the four
// things that can be done to the song, or for a link the one that sends its songs on. They sit on
// the last row of the details pane. None when there is nothing to choose or no room for them.
func (m Model) choiceButtons() []choiceButton {
	if _, held := m.chooser(); !held || m.technical || m.fullLog || m.adding || m.swapping || m.showHelp {
		return nil
	}
	l, ok := m.layout()
	if !ok || !l.hasOutput {
		return nil
	}
	var labels [][2]string
	switch {
	case m.focus == paneLists:
		if m.selPlaylist >= len(m.state.Playlists) || m.waitingIn(m.state.Playlists[m.selPlaylist]) == 0 {
			return nil
		}
		labels = [][2]string{{"► Download these songs", "d"}}
	case waiting(m.selectedEntry()):
		labels = [][2]string{{"► Download", "d"}, {"⇄ Swap video", "s"}, {"✗ Cancel song", "x"}, {"↗ Open video", "o"}}
	case swappable(m.selectedEntry()):
		labels = [][2]string{{"⇄ Swap video", "s"}, {"↗ Open video", "o"}}
	default:
		return nil
	}

	inner := l.output.rows()
	x, y := inner.x+1, inner.y+inner.h-1
	var out []choiceButton
	for _, label := range labels {
		w := len([]rune(label[0])) + 4
		if x+w > inner.x+inner.w {
			break
		}
		out = append(out, choiceButton{label: label[0], key: label[1], x: x, y: y, w: w})
		x += w + 2
	}

	return out
}

// withChoiceButtons draws the buttons over the screen.
func (m Model) withChoiceButtons(screen string) string {
	for i, b := range m.choiceButtons() {
		kind := dialog.PillNormal
		if i == 0 {
			kind = dialog.PillPrimary
		}
		screen = dialog.Overlay(screen, dialog.WidePill(b.label, b.w, kind, false), b.x, b.y)
	}

	return screen
}

// clickChoice does what the button under (x, y) does, when there is one.
func (m Model) clickChoice(x, y int) (Model, tea.Cmd, bool) {
	for _, b := range m.choiceButtons() {
		if y == b.y && x >= b.x && x < b.x+b.w {
			if key, ok := keybar.KeyMsg(b.key); ok {
				next, cmd := m.handleKey(key)

				return next.(Model), cmd, true
			}
		}
	}

	return m, nil, false
}
