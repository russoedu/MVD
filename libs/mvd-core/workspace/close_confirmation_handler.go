package workspace

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// CloseRequestedMsg tells the workspace that the person asked to close the app, for example by the
// close button of its window. The workspace closes it at once, unless songs are being downloaded,
// when it asks first.
type CloseRequestedMsg struct{}

// closeAsk is the question asked before the app closes while songs download: which of its two
// buttons has the focus (0 quits, 1 keeps downloading).
type closeAsk struct {
	on    bool
	focus int
}

// downloading is how many songs are being downloaded or wait for their turn to be.
func (m Model) downloading() int {
	if m.run == nil {
		return 0
	}
	t := m.screen.State().Tally()
	if m.run.Paused() {
		return t.Running // what was accepted waits for the Download button: nothing is lost by quitting
	}

	return t.Accepted + t.Running
}

// closeRequested closes the app, after asking when songs are being downloaded.
func (m Model) closeRequested() (Model, tea.Cmd) {
	if m.downloading() > 0 {
		m.closing = closeAsk{on: true, focus: 1}

		return m, nil
	}

	return m.stop(true)
}

var closeButtons = []string{"Quit", "Keep downloading"}

// closeLayout is the question drawn: its lines, the buttons on the last, and the size of the frame.
func (m Model) closeLayout() (lines []string, buttons []dialog.Span, w, h int) {
	w = min(max(m.width*60/100, 46), max(m.width-2, 20))
	n := m.downloading()
	what := fmt.Sprintf("%d songs are still downloading.", n)
	if n == 1 {
		what = "1 song is still downloading."
	}
	lines = []string{
		"",
		"  " + look.Text.Render(what),
		"  " + look.Text.Render("Quit and stop them?"),
		"",
	}
	label, spans := dialog.Buttons(closeButtons, m.closing.focus, nil)
	pad := max((w-2-spans[len(spans)-1].X1)/2, 0)
	lines = append(lines, strings.Repeat(" ", pad)+label)
	for i := range spans {
		spans[i] = dialog.Span{X0: spans[i].X0 + pad, X1: spans[i].X1 + pad}
	}

	return lines, spans, w, len(lines) + 2
}

// closeView is the question, as a frame.
func (m Model) closeView() string {
	lines, _, w, h := m.closeLayout()

	return dialog.Frame("Quit MVD?", lines, w, h)
}

// closeOrigin is the cell of the screen where the question begins.
func (m Model) closeOrigin() (int, int) {
	_, _, w, h := m.closeLayout()

	return dialog.Centre(m.width, m.height, w, h)
}

// closeKey works a key while the question is asked.
func (m Model) closeKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "y", "Y":
		return m.quitNow()
	case "n", "N", "esc":
		m.closing = closeAsk{}
	case "left", "right", "tab", "shift+tab", "h", "l":
		m.closing.focus = 1 - m.closing.focus
	case "enter", "space":
		if m.closing.focus == 0 {
			return m.quitNow()
		}
		m.closing = closeAsk{}
	}

	return m, nil
}

// quitNow stops the downloads and closes the app.
func (m Model) quitNow() (Model, tea.Cmd) {
	m.closing = closeAsk{}

	return m.stop(true)
}

// closeMouse works the mouse while the question is asked: a click on a button answers it, and
// the pointer over one gives it the focus.
func (m Model) closeMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	_, spans, _, _ := m.closeLayout()
	x0, y0 := m.closeOrigin()
	pos := msg.Mouse()
	row := pos.Y == y0+1+4 // the buttons are on the fifth line inside the frame
	for i, s := range spans {
		if !row || !s.Contains(pos.X-x0-1) {
			continue
		}
		switch msg.(type) {
		case tea.MouseMotionMsg:
			m.closing.focus = i
		case tea.MouseClickMsg:
			if pos.Button != tea.MouseLeft {
				return m, nil
			}
			if i == 0 {
				return m.quitNow()
			}
			m.closing = closeAsk{}
		}
	}

	return m, nil
}
