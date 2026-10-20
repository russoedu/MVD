package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	advancedpage "youtube-downloader/libs/mvd-core/setup/advanced"
	colourspage "youtube-downloader/libs/mvd-core/setup/colours"
)

// The popup of the preferences: a row of tabs, the screen of the tab under it, and a row
// of buttons. The screens are the ones the preferences have always had; the tabs and the
// buttons are the popup's own.
var (
	popupTabs    = []string{"Preferences", "Advanced", "Colours"}
	popupButtons = []string{"Save", "Close"}
)

const (
	tabPreferences = iota
	tabAdvanced
	tabColours
)

const (
	buttonSave = iota
	buttonClose
)

// activeTab is the tab whose screen is showing.
func (m setupModel) activeTab() int {
	switch m.screen {
	case screenAdvanced:
		return tabAdvanced
	case screenColours:
		return tabColours
	}
	return tabPreferences
}

// commit takes what the screen showing has changed into the preferences.
func (m setupModel) commit() setupModel {
	switch m.screen {
	case screenAdvanced:
		m.config = m.config.WithConfig(m.advanced.Config())
	case screenColours:
		m.config = m.config.WithConfig(m.colours.Config())
	}
	return m
}

// openTab shows the screen of tab, keeping the changes made in the one it leaves.
func (m setupModel) openTab(tab int) setupModel {
	if tab == m.activeTab() {
		return m
	}
	m = m.commit()
	switch tab {
	case tabAdvanced:
		m.advanced = advancedpage.New(m.config.Config()).SetSize(m.width, m.height)
		m.screen = screenAdvanced
	case tabColours:
		m.colours = colourspage.New(m.config.Config()).SetSize(m.width, m.height)
		m.screen = screenColours
	default:
		m.screen = screenConfig
	}
	return m
}

// bodySize is the room the screen of a tab takes: the inside of the frame less the rows
// of the tabs and of the buttons.
func (a workspaceApp) bodySize() (int, int) {
	w, h := a.popupSize()
	return w, h - 2
}

// popupLines are the lines inside the frame: tabs, the screen, buttons.
func (a workspaceApp) popupLines() []string {
	_, h := a.popupSize()
	body := h - 4
	tabs, _ := dialog.Tabs(popupTabs, a.prefs.activeTab())
	lines := []string{" " + tabs}
	screen := strings.Split(a.prefs.render(), "\n")
	for i := 0; i < body; i++ {
		if i < len(screen) {
			lines = append(lines, screen[i])
		} else {
			lines = append(lines, "")
		}
	}
	buttons, _ := dialog.Buttons(popupButtons, -1, nil)
	return append(lines, " "+buttons)
}

// inPopup puts a mouse event in the coordinates of the inside of the frame, and says
// whether it was inside: the frame, and what is outside it, take no clicks.
func (a workspaceApp) inPopup(msg tea.MouseMsg) (tea.MouseMsg, bool) {
	w, h := a.popupSize()
	x, y := dialog.Centre(a.width, a.height, w, h)
	m := msg.Mouse()
	m.X, m.Y = m.X-x-1, m.Y-y-1
	if m.X < 0 || m.Y < 0 || m.X >= w-2 || m.Y >= h-2 {
		return nil, false
	}
	return withMouse(msg, m)
}

// withMouse is msg of the same kind as it was, with m for where it happened.
func withMouse(msg tea.MouseMsg, m tea.Mouse) (tea.MouseMsg, bool) {
	switch msg.(type) {
	case tea.MouseClickMsg:
		return tea.MouseClickMsg(m), true
	case tea.MouseReleaseMsg:
		return tea.MouseReleaseMsg(m), true
	case tea.MouseMotionMsg:
		return tea.MouseMotionMsg(m), true
	case tea.MouseWheelMsg:
		return tea.MouseWheelMsg(m), true
	}
	return nil, false
}

// popupMouse works out an event inside the frame: a click on a tab or a button is the
// popup's; any other goes to the screen under the tabs, in its own coordinates.
func (a workspaceApp) popupMouse(msg tea.MouseMsg) (workspaceApp, tea.Msg, bool) {
	m := msg.Mouse()
	_, h := a.popupSize()
	if _, ok := msg.(tea.MouseClickMsg); ok && m.Button == tea.MouseLeft {
		switch m.Y {
		case 0:
			_, spans := dialog.Tabs(popupTabs, a.prefs.activeTab())
			if i := dialog.At(spans, m.X-1); i >= 0 {
				a.prefs = a.prefs.openTab(i)
			}
			return a, nil, false
		case h - 3:
			_, spans := dialog.Buttons(popupButtons, -1, nil)
			switch dialog.At(spans, m.X-1) {
			case buttonSave:
				a.prefs = a.prefs.commit()
				a = a.apply(a.prefs.config.Config(), true)
				a.inPrefs = false
				ApplyTheme(a.cfg.Colors)
			case buttonClose:
				a.inPrefs = false
				ApplyTheme(a.cfg.Colors)
			}
			return a, nil, false
		}
	}
	if m.Y == 0 || m.Y == h-3 {
		return a, nil, false
	}
	m.Y--
	inner, ok := withMouse(msg, m)
	return a, inner, ok
}
