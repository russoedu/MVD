package workspace

import (
	"youtube-downloader/libs/mvd-core/workspace/closequestion"
	"strings"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/look"
)

// CloseRequestedMsg tells the workspace that the person asked to close the app, for example by the
// close button of its window. The workspace closes it at once, unless songs are being downloaded,
// when it asks first.
type CloseRequestedMsg struct{}

// downloading is how many songs are being downloaded or wait for their turn to be.
func (m Model) downloading() int { return m.facts().Downloading() }

// closeRequested closes the app, after asking when songs are being downloaded.
func (m Model) closeRequested() (Model, tea.Cmd) {
	if m.downloading() > 0 {
		m.closing = closequestion.Ask()

		return m, nil
	}

	return m.stop(true)
}

// closeLayout is the question drawn: its lines, the buttons on the last, and the size of the frame.
func (m Model) closeLayout() (lines []string, buttons []dialog.Span, w, h int) {
	w = min(max(m.width*60/100, 46), max(m.width-2, 20))
	what, ask := closequestion.Words(m.downloading())
	lines = []string{
		"",
		"  " + look.Text.Render(what),
		"  " + look.Text.Render(ask),
		"",
	}
	label, spans := dialog.Buttons(closequestion.Answers, m.closing.Focus, nil)
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
	closing, outcome := m.closing.Key(key.String())
	m.closing = closing
	if outcome == closequestion.QuitNow {
		return m.quitNow()
	}

	return m, nil
}

// quitNow stops the downloads and closes the app.
func (m Model) quitNow() (Model, tea.Cmd) {
	m.closing = closequestion.State{}

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
			m.closing = m.closing.Hover(i)
		case tea.MouseClickMsg:
			if pos.Button != tea.MouseLeft {
				return m, nil
			}
			closing, outcome := m.closing.Click(i)
			m.closing = closing
			if outcome == closequestion.QuitNow {
				return m.quitNow()
			}
		}
	}

	return m, nil
}
