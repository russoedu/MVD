package workspace

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/menubar"
	"youtube-downloader/libs/mvd-core/pointer"
)

// popupPercent is how much of the screen the edit of the list takes, in each direction.
const popupPercent = 85

// editSize is the room the edit of the list takes: most of the screen, whatever its size.
func (m Model) editSize() (int, int) {
	return dialog.Sized(m.width, m.height, popupPercent, 40, 12)
}

// editOrigin is the cell of the screen where the edit of the list begins.
func (m Model) editOrigin() (int, int) {
	w, h := m.editSize()
	return dialog.Centre(m.width, m.height, w, h)
}

// aboutOrigin is the cell of the screen where the About window begins.
func (m Model) aboutOrigin() (int, int) {
	l := m.aboutLayout()
	return dialog.Centre(m.width, m.height, l.width, l.height)
}

// popupMouse gives a mouse event to the popup that is open, if there is one. A click on the
// menu bar is left to the bar, which is over the popups.
func (m Model) popupMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	click, isClick := msg.(tea.MouseClickMsg)
	if isClick && click.Y == 0 {
		return m, nil, false
	}
	switch {
	case m.about:
		return m.aboutMouse(msg)
	case m.editing:
		x, y := m.editOrigin()
		w, h := m.editSize()
		pos := msg.Mouse()
		inside := pos.X >= x && pos.X < x+w && pos.Y >= y && pos.Y < y+h
		if !inside && !m.edit.dragging {
			return m, nil, true
		}
		m.edit = m.edit.sized(w, h)
		var cmd tea.Cmd
		m.edit, cmd = m.edit.mouse(pointer.Shift(msg, -x, -y))
		return m, cmd, true
	}
	return m, nil, false
}

// aboutMouse works out a click on the About window: its close mark and its button close it,
// the address opens in the browser.
func (m Model) aboutMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return m, nil, true
	}
	l := m.aboutLayout()
	x0, y0 := m.aboutOrigin()
	x, y := click.X-x0, click.Y-y0
	switch {
	case y == 0 && dialog.CloseSpan(l.width).Contains(x):
		m.about = false
	case y-1 == min(l.buttonY, l.height-3) && l.button.Contains(x-1):
		m.about = false
	case y-1 == l.linkY && l.link.Contains(x-1):
		if m.host.Open != nil {
			if err := m.host.Open(projectURL); err != nil {
				m.notice = "Cannot open the browser: " + err.Error()
			}
		}
	}
	return m, nil, true
}

// refreshMenus greys out what the menus offer and the screen cannot do now: changing the
// list is done in the edit of the list, so its commands wait for that to be open.
func (m Model) refreshMenus() Model {
	for _, name := range []string{entryUndo, entryRedo, entryCopy, entryPaste, entryDelete, entryMoveUp, entryMoveDown} {
		m.bar.SetEnabled(name, m.editing)
	}
	return m
}

// openEditFirst is what to say when a command of the edit of the list is used with it closed.
func openEditFirst(name string) string {
	return "Open Edit list first (" + menubar.ShortcutLabel("Ctrl+E") + "), then " + strings.ToLower(name) + " there"
}
