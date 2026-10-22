package workspace

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
)

// topRows is the rows above the screen: the menu bar.
const topRows = 1

// The buttons of the toolbar, and the Run menu entries that do the same.
const (
	entryStart       = "Start"
	entryReview      = "Get data for review"
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
		return []button{
			{"▶ Start", entryStart, busy},
			{"Get data for review", entryReview, busy},
			{"+ Add link", entryAddLink, busy},
		}
	}
	pause := button{"⏸ Pause", entryPause, busy}
	if m.run.Paused() {
		pause = button{"▶ Resume", entryResume, busy}
	}
	out := []button{pause, {"■ Stop", entryStop, busy}, {"↻ Retry failed", entryRetry, busy}, {"+ Add link", entryAddLink, busy}}
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
	switch {
	case buttons[i].disabled:
		return dialog.PillDisabled
	case i == 0:
		return dialog.PillPrimary
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
		screen = dialog.Overlay(screen, dialog.WidePill(b.label, w, buttonKind(buttons, i)), x, y+i)
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
	m.bar.SetEnabled(entryStart, !running && !busy)
	m.bar.SetEnabled(entryReview, !running && !busy)
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
