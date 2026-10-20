package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	preferencespage "youtube-downloader/libs/mvd-core/setup/preferences"
)

// The popup of the preferences is the prefspopup, in a frame, with a row of buttons under
// it. The tabs, and the settings under them, are the prefspopup's own.
var popupButtons = []string{"Save", "Close"}

const (
	buttonSave = iota
	buttonClose
)

// bodySize is the room the prefspopup takes: the inside of the frame less the row of
// buttons.
func (a workspaceApp) bodySize() (int, int) {
	w, h := a.popupSize()
	return max(w-2, 1), max(h-3, 1)
}

// popupLines are the lines inside the frame: the prefspopup, then the buttons.
func (a workspaceApp) popupLines() []string {
	lines := strings.Split(a.prefs.View(), "\n")
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

// popupMouse works out an event inside the frame: a click on a button is the app's; any
// other goes to the prefspopup, which takes the same coordinates.
func (a workspaceApp) popupMouse(msg tea.MouseMsg) (workspaceApp, tea.Msg, bool) {
	m := msg.Mouse()
	_, h := a.bodySize()
	if m.Y < h {
		return a, msg, true
	}
	if _, ok := msg.(tea.MouseClickMsg); ok && m.Button == tea.MouseLeft {
		_, spans := dialog.Buttons(popupButtons, -1, nil)
		switch dialog.At(spans, m.X-1) {
		case buttonSave:
			a = a.closePrefs(true)
		case buttonClose:
			a = a.closePrefs(false)
		}
	}
	return a, nil, false
}

// movingResult is what to tell the person about the move of the app.
func movingResult(outcome preferencespage.MoveOutcome, err error) (string, error) {
	switch {
	case err != nil:
		return "", err
	case outcome == preferencespage.MoveStarted:
		return "MVD was moved to its own folder. This window will close and the new copy will open.", nil
	case outcome == preferencespage.MoveAlreadyThere:
		return "MVD already lives in its own folder.", nil
	case outcome == preferencespage.MoveUnavailable:
		return "This machine has no place to move MVD to, or no way to ask. Start MVD with -move from a terminal.", nil
	}
	return "MVD was not moved.", nil
}

// removalResult is what to tell the person about the removal of the app.
func removalResult(outcome preferencespage.RemovalOutcome, err error) (string, error) {
	switch {
	case err != nil:
		return "", err
	case outcome == preferencespage.RemovalStarted:
		return "Removing MVD. This window will close.", nil
	case outcome == preferencespage.RemovalDeclined:
		return "Nothing was removed.", nil
	}
	return "This machine cannot ask for confirmation, so nothing was removed. Start MVD with -uninstall from a terminal.", nil
}
