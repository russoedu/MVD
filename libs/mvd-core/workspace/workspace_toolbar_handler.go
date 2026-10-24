package workspace

import (
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
)

// topRows is the rows above the screen: the menu bar.
const topRows = 1

// The buttons of the toolbar, and the Run menu entries that do the same.
const (
	entryStart       = "Start"
	entryReview      = "Look up songs first"
	entryAddLink     = "Add link…"
	entryPause       = "Pause"
	entryResume      = "Resume"
	entryStop        = "Stop"
	entryRetry       = "Retry failed"
	entryReviewSongs = "Review songs"
)

// button is one button of the toolbar.
type button struct {
	label    string // what it says
	entry    string // what choosing it does, as the entry of a menu is named
	disabled bool
}

// buttons are the toolbar's buttons for what the workspace is doing: the list can be started
// or edited while nothing runs; a run can be paused, stopped, retried and reviewed.
func (m Model) buttons() []button {
	busy := m.starting || m.stopping
	if !m.running() {
		_, waiting := m.toolsWaiting()
		nothing := busy || waiting || len(m.session.Sources()) == 0 // there is nothing to download, or no tool to do it with
		return []button{
			{"► Start", entryStart, nothing},
			{"Look up first", entryReview, nothing},
			{"+ Add link", entryAddLink, busy},
		}
	}
	pause := button{"Pause", entryPause, busy}
	if m.run.Paused() {
		pause = button{"► Resume", entryResume, busy}
	}
	out := []button{pause, {"■ Stop", entryStop, busy}, {"Retry failed", entryRetry, busy}, {"+ Add link", entryAddLink, busy}}
	if m.editor.Live() {
		out = append(out, button{"Review songs", entryReviewSongs, busy})
	}
	return out
}

// buttonRows is where the buttons are: the rows at the bottom of the column of playlists,
// counted on the screen, with the left column and the width of the buttons. ok is false
// when the screen has no such column (a narrow window), and the Run menu is the way then.
func (m Model) buttonRows() (x, y, w int, ok bool) {
	cx, cy, cw, ch, reserve, drawn := m.screen.LeftColumn()
	if !drawn || reserve == 0 || m.reviewing {
		return 0, 0, 0, false
	}
	return cx, cy + (ch - reserve) + topRows, cw, true
}

// buttonKind is how a button looks: the first one that can be used is the main one.
func buttonKind(buttons []button, i int) dialog.PillKind {
	if buttons[i].disabled {
		return dialog.PillDisabled
	}
	for j, b := range buttons {
		if !b.disabled {
			if j == i {
				return dialog.PillPrimary
			}
			break
		}
	}
	return dialog.PillNormal
}

// buttonsView draws the buttons in the rows kept for them.
func (m Model) buttonsView(screen string) string {
	x, y, w, ok := m.buttonRows()
	if !ok {
		return screen
	}
	buttons := m.buttons()
	for i, b := range buttons {
		hover := m.hoverY == y+i && m.hoverX >= x && m.hoverX < x+w
		screen = dialog.Overlay(screen, dialog.WidePill(b.label, w, buttonKind(buttons, i), hover), x, y+i)
	}
	return screen
}

// toolbarClick does what the button at (x, y) of the screen does; it reports false when
// there is none.
func (m Model) toolbarClick(x, y int) (tea.Model, tea.Cmd, bool) {
	bx, by, bw, ok := m.buttonRows()
	if !ok || x < bx || x >= bx+bw {
		return m, nil, false
	}
	buttons := m.buttons()
	i := y - by
	if i < 0 || i >= len(buttons) {
		return m, nil, false
	}
	if buttons[i].disabled {
		return m, nil, true
	}
	next, cmd := m.choose(buttons[i].entry)
	return next, cmd, true
}

// refreshRunMenus greys out what the Run menu offers and the workspace cannot do now.
func (m Model) refreshRunMenus() Model {
	busy := m.starting || m.stopping
	running := m.running()
	nothing := len(m.session.Sources()) == 0
	_, waiting := m.toolsWaiting()
	m.bar.SetEnabled(entryStart, !running && !busy && !nothing && !waiting)
	m.bar.SetEnabled(entryReview, !running && !busy && !nothing && !waiting)
	m.bar.SetEnabled(entryPause, running && !busy && !m.run.Paused())
	m.bar.SetEnabled(entryResume, running && !busy && m.run.Paused())
	m.bar.SetEnabled(entryStop, running && !busy)
	m.bar.SetEnabled(entryRetry, running && !busy)
	m.bar.SetEnabled(entryReviewSongs, running && !busy && m.editor.Live())
	return m
}

// runChoice does what the entries of the Run menu, and the buttons of the toolbar, say.
func (m Model) runChoice(name string) (tea.Model, tea.Cmd, bool) {
	switch name {
	case entryStart:
		next, cmd := m.start(false)
		return next, cmd, true
	case entryReview:
		next, cmd := m.start(true)
		return next, cmd, true
	case entryPause, entryResume:
		return m.togglePause(), nil, true
	case entryStop:
		next, cmd := m.stop(false)
		return next, cmd, true
	case entryRetry:
		return m.retryAll(), nil, true
	case entryAddLink:
		if m.editable() {
			next, cmd := m.startAdd()
			return next, cmd, true
		}
		return m.ask(askAddLink, "Add a playlist or video link", ""), textinput.Blink, true
	case entryReviewSongs:
		if m.running() && m.editor.Live() {
			m.reviewing = true
		}
		return m, nil, true
	}
	return m, nil, false
}

// keepRoomForButtons tells the screen how many rows under the playlists the buttons take, and
// whether the list can be changed on it.
func (m Model) keepRoomForButtons() Model {
	m.screen = m.screen.WithReserve(len(m.buttons())).WithEditable(m.editable())
	return m
}

// disabledReason is why the button under the pointer cannot be used, or "" when the pointer is
// not over a button that cannot.
func (m Model) disabledReason() string {
	x, y, w, ok := m.buttonRows()
	if !ok {
		return ""
	}
	for i, b := range m.buttons() {
		if b.disabled && m.hoverY == y+i && m.hoverX >= x && m.hoverX < x+w {
			if m.starting || m.stopping {
				return "One moment..."
			}
			if state, waiting := m.toolsWaiting(); waiting {
				return state.Message
			}
			return "Add a link first."
		}
	}
	return ""
}

// toolsWaiting says whether the tools of the downloads are not ready, and what they are.
func (m Model) toolsWaiting() (ToolsState, bool) {
	if m.host.Tools == nil {
		return ToolsState{}, false
	}
	state := m.host.Tools()

	return state, !state.Ready
}

// toolsEvery is how often the state of the tools is looked at while they are being fetched.
const toolsEvery = 400 * time.Millisecond

// toolsTick asks to be woken to look at the tools again, as long as they are being fetched; it
// is nothing once they are ready or when one could not be gotten, which does not change.
func (m Model) toolsTick() tea.Cmd {
	state, waiting := m.toolsWaiting()
	if !waiting || state.Failed {
		return nil
	}

	return tea.Tick(toolsEvery, func(t time.Time) tea.Msg { return toolsTickMsg(t) })
}
