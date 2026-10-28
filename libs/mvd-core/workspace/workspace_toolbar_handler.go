package workspace

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/workspace/runcontrol"
)

// topRows is the rows above the screen: the menu bar.
const topRows = 1

// The buttons of the toolbar, and the Run menu entries that do the same.
const (
	entryStart       = "Load playlists"
	entryAddLink     = "Add source…"
	entryPause       = "Pause downloads"
	entryResume      = "Download"
	entryStop        = "Stop and cancel"
	entryRetry       = "Retry failed"
	entryDownloadAll = "Accept all songs"
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
	busy := m.life.Busy()
	if m.finished() {
		// The work is over: show where it went, and offer what comes next.
		out := []button{{"Open folder", entryOpenFolder, false}, {"► Load playlists again", entryStart, busy}}
		if m.dl.State.Tally().Failed > 0 {
			out = append(out, button{"Retry failed", entryRetry, busy})
		}
		out = append(out, button{"+ Add source", entryAddLink, busy})
		if m.dl.State.Tally().Done == 0 {
			// Nothing was saved, so the folder is not where the person goes next: the first button
			// that can be used is the main one, and that is the way to give another source.
			out = append(out[len(out)-1:], out[:len(out)-1]...)
		}

		return out
	}
	if !m.running() {
		_, waiting := m.toolsWaiting()
		nothing := busy || waiting || len(m.session.Sources()) == 0 // there is nothing to download, or no tool to do it with
		return []button{
			{"► Load playlists", entryStart, nothing},
			{"+ Add source", entryAddLink, busy},
		}
	}
	// Download starts the downloads of what was accepted, and Pause downloads holds them again: the
	// songs that are on their way finish, and nothing else begins.
	download := button{"► Download", entryResume, busy}
	if !m.run.Paused() {
		download = button{"Pause downloads", entryPause, busy}
	}
	out := []button{download, {"Retry failed", entryRetry, busy || m.facts().Tally.Failed == 0}, {"+ Add source", entryAddLink, busy}}
	return out
}

// buttonRows is where the buttons are: the rows at the bottom of the column of playlists,
// counted on the screen, with the left column and the width of the buttons. ok is false
// when the screen has no such column (a narrow window), and the Run menu is the way then.
func (m Model) buttonRows() (x, y, w int, ok bool) {
	cx, cy, cw, ch, reserve, drawn := m.view().LeftColumn()
	if !drawn || reserve == 0 {
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
	_, waiting := m.toolsWaiting()
	menu := runcontrol.RunMenuOf(m.facts(), len(m.session.Sources()) == 0, waiting)
	m.bar.SetEnabled(entryStart, menu.Start)
	m.bar.SetEnabled(entryPause, menu.Pause)
	m.bar.SetEnabled(entryResume, menu.Resume)
	m.bar.SetEnabled(entryStop, menu.Stop)
	m.bar.SetEnabled(entryRetry, menu.Retry)
	m.bar.SetEnabled(entryOpenFolder, m.host.OpenFolder != nil)
	m.bar.SetEnabled(entryDownloadAll, menu.DownloadAll)
	return m
}

// runChoice does what the entries of the Run menu, and the buttons of the toolbar, say.
func (m Model) runChoice(name string) (tea.Model, tea.Cmd, bool) {
	switch name {
	case entryStart:
		next, cmd := m.start(true)
		return next, cmd, true
	case entryPause, entryResume:
		return m.togglePause(), nil, true
	case entryStop:
		next, cmd := m.stop(false)
		return next, cmd, true
	case entryRetry:
		return m.retryAll(), nil, true
	case entryOpenFolder:
		m.openFolder()
		return m, nil, true
	case entryAddLink:
		next, cmd := m.openAdd()
		return next, cmd, true
	case entryDownloadAll:
		if m.running() {
			m.done = runcontrol.AcceptedNotice(m.run.ReleaseAll())
		}
		return m, nil, true
	}
	return m, nil, false
}

// keepRoomForButtons tells the screen how many rows under the playlists the buttons take, and
// whether the list can be changed on it.
func (m Model) keepRoomForButtons() Model {
	m.show(m.view().WithReserve(len(m.buttons())).WithEditable(m.editable()).WithLinkEditing(m.addable()))
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
			state, waiting := m.toolsWaiting()
			return runcontrol.DisabledReason(m.facts(), waiting, state.Message)
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
