package workspace

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"youtube-downloader/libs/mvd-core/cells"
	"youtube-downloader/libs/mvd-core/dialog"
)

// topRows is the rows above the screen: the menu bar, and the toolbar under it.
const topRows = 2

// toolbarRow is the row of the toolbar.
const toolbarRow = 1

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
			{"✎ Edit list", entryEditList, busy},
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

// toolbarLine draws the toolbar and says where each button is, counted from the left of the
// screen. The first button, the one the person most likely wants, is lit.
func (m Model) toolbarLine() (string, []dialog.Span) {
	buttons := m.buttons()
	labels := make([]string, len(buttons))
	disabled := map[string]bool{}
	for i, b := range buttons {
		labels[i] = b.label
		disabled[b.label] = b.disabled
	}
	focus := 0
	if buttons[0].disabled {
		focus = -1
	}
	line, spans := dialog.Buttons(labels, focus, disabled)
	for i := range spans {
		spans[i].X0++
		spans[i].X1++
	}
	return cells.Fit(" "+line, m.width), spans
}

// toolbarClick does what the button under column x does; it reports false when there is none.
func (m Model) toolbarClick(x int) (tea.Model, tea.Cmd, bool) {
	_, spans := m.toolbarLine()
	i := dialog.At(spans, x)
	if i < 0 {
		return m, nil, false
	}
	b := m.buttons()[i]
	if b.disabled {
		return m, nil, true
	}
	next, cmd := m.choose(b.entry)
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
	m.bar.SetEnabled(entryEditList, !running && !busy)
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

// toolbarView is the row drawn under the menu bar.
func (m Model) toolbarView() string {
	line, _ := m.toolbarLine()
	return strings.TrimRight(line, " ") + strings.Repeat(" ", max(m.width-ansi.StringWidth(strings.TrimRight(line, " ")), 0))
}
