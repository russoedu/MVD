package workspace

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
)

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

// refreshMenus greys out what the menus offer and the screen cannot do now: the list is
// changed in place, so its commands wait for the downloads to stop.
func (m Model) refreshMenus() Model {
	for _, name := range []string{entryUndo, entryRedo, entryCopy, entryPaste, entryDelete, entryMoveUp, entryMoveDown} {
		m.bar.SetEnabled(name, m.editable() && !m.reviewing)
	}
	m.bar.SetEnabled(entryUndo, m.editable() && !m.reviewing && len(m.undone) > 0)
	m.bar.SetEnabled(entryRedo, m.editable() && !m.reviewing && len(m.redone) > 0)
	return m
}
