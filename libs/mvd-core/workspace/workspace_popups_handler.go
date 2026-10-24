package workspace

import (
	"net/url"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
)

// AboutOpen reports whether the About window is showing.
func (m Model) AboutOpen() bool { return m.about }

// Confirm shows, in green at the bottom of the screen, that what the person asked for was done.
func (m Model) Confirm(text string) Model {
	m.done, m.caution = text, ""
	return m
}

// Caution shows, in yellow, something the person should know about what they just did.
func (m Model) Caution(text string) Model {
	m.caution, m.done = text, ""
	return m
}

// WindowOpen reports whether a window of the workspace (About, Keyboard shortcuts) is showing
// over the screen, which the page around the terminal dims itself for.
func (m Model) WindowOpen() bool { return m.about || m.shortcuts }

// shortcutsKey works a key press while the Keyboard shortcuts window is showing: it scrolls,
// or it closes.
func (m Model) shortcutsKey(key tea.KeyPressMsg) Model {
	l := m.shortcutsLayout()
	page := max(l.height-4, 1)
	switch key.String() {
	case "esc", "enter", "space", "q", "f1", "ctrl+k":
		m.shortcuts = false
	case "up", "k":
		m.shortcutsScroll = max(m.shortcutsScroll-1, 0)
	case "down", "j":
		m.shortcutsScroll = min(m.shortcutsScroll+1, l.maxScroll)
	case "pgup":
		m.shortcutsScroll = max(m.shortcutsScroll-page, 0)
	case "pgdown":
		m.shortcutsScroll = min(m.shortcutsScroll+page, l.maxScroll)
	}
	return m
}

// shortcutsOrigin is the cell of the screen where the Keyboard shortcuts window begins.
func (m Model) shortcutsOrigin() (int, int) {
	l := m.shortcutsLayout()
	return dialog.Centre(m.width, m.height, l.width, l.height)
}

// shortcutsMouse works out a click on the Keyboard shortcuts window: its close mark closes it,
// and the wheel scrolls it.
func (m Model) shortcutsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	l := m.shortcutsLayout()
	switch ev := msg.(type) {
	case tea.MouseWheelMsg:
		switch ev.Button {
		case tea.MouseWheelUp:
			m.shortcutsScroll = max(m.shortcutsScroll-2, 0)
		case tea.MouseWheelDown:
			m.shortcutsScroll = min(m.shortcutsScroll+2, l.maxScroll)
		}
	case tea.MouseClickMsg:
		x0, y0 := m.shortcutsOrigin()
		if ev.Button == tea.MouseLeft && ev.Y-y0 == 0 && dialog.CloseSpan(l.width).Contains(ev.X-x0) {
			m.shortcuts = false
		}
	}
	return m, nil, true
}

// openAddress opens an address in the person's browser, or says it when there is none.
func (m *Model) openAddress(address string) {
	if m.host.Open == nil {
		m.notice = "Open " + address + " in your browser"
		return
	}
	if err := m.host.Open(address); err != nil {
		m.notice = "Cannot open the browser: " + err.Error()
	}
}

// reportAddress is the page for a new issue of the project, with the version already in it.
func (m Model) reportAddress() string {
	version := m.host.Version
	if version == "" {
		version = "dev"
	}
	body := "MVD " + version + "\n\nWhat I did:\n\nWhat I expected:\n\nWhat happened instead:\n"
	return projectURL + "/issues/new?body=" + url.QueryEscape(body)
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
	case m.shortcuts:
		return m.shortcutsMouse(msg)
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

// refreshMenus greys out what the menus offer and does not make sense now. Copy and Paste are
// for the address being written, and Copy needs something selected; the other commands work
// on the playlist the cursor is on, and wait for the downloads to stop.
func (m Model) refreshMenus() Model {
	free := m.editable() && !m.reviewing
	writing := free && m.inline.on
	i, has := m.selected()
	n := len(m.session.File().Items)
	m.bar.SetEnabled(entryCut, writing && m.inline.line.HasSelection())
	m.bar.SetEnabled(entryCopy, writing && m.inline.line.HasSelection())
	m.bar.SetEnabled(entryPaste, free)
	m.bar.SetEnabled(entryUndo, free && !writing && len(m.undone) > 0)
	m.bar.SetEnabled(entryRedo, free && !writing && len(m.redone) > 0)
	m.bar.SetEnabled(entryDelete, free && !writing && has)
	m.bar.SetEnabled(entryMoveUp, free && !writing && has && i > 0)
	m.bar.SetEnabled(entryMoveDown, free && !writing && has && i < n-1)
	return m
}

// Background says whether a message is the workspace's own beat, which it must get whatever popup
// is open over it: the one that keeps the list up to date with a run, and the one that looks at
// the tools again while they are being fetched.
func Background(msg tea.Msg) bool {
	switch msg.(type) {
	case saveTickMsg, toolsTickMsg:
		return true
	}

	return false
}
