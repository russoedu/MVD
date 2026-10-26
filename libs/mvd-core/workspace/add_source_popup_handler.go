package workspace

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/dialog"
	"youtube-downloader/libs/mvd-core/workspace/addsource"
)

// openAdd shows the window for adding music, with the platform address chosen.
func (m Model) openAdd() (Model, tea.Cmd) {
	m = m.stopInline(false)
	m.add, m.addOpen = addsource.New(), true
	for f := addsource.FieldURL; f <= addsource.FieldSong; f++ {
		in := textinput.New()
		in.Prompt = ""
		in.SetWidth(max(m.addLayoutBoxWidth(), 8))
		m.addBoxes[f] = in
	}

	return m.syncAddFocus(), textinput.Blink
}

// addLayoutBoxWidth is how wide a text field of the window is.
func (m Model) addLayoutBoxWidth() int { return max(m.width-2-2-labelCells-2, 8) - 2 }

// syncAddFocus gives the text field the focus is on the cursor, and the others none.
func (m Model) syncAddFocus() Model {
	for f := addsource.FieldURL; f <= addsource.FieldSong; f++ {
		in := m.addBoxes[f]
		if m.add.Focus == f {
			in.Focus()
		} else {
			in.Blur()
		}
		m.addBoxes[f] = in
	}

	return m
}

// closeAdd puts the window away.
func (m Model) closeAdd() Model {
	m.addOpen = false

	return m
}

// addKey works a key press while the window for adding music is showing.
func (m Model) addKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := key.String()
	switch name {
	case "esc":
		return m.closeAdd(), nil
	case "tab", "down":
		m.add = m.add.Next()

		return m.syncAddFocus(), nil
	case "shift+tab", "up":
		m.add = m.add.Previous()

		return m.syncAddFocus(), nil
	}
	switch focus := m.add.Focus; {
	case focus == addsource.FieldMode:
		switch name {
		case "left", "right", "h", "l", "space":
			m.add = m.add.WithMode(1 - m.add.Mode)

			return m.syncAddFocus(), nil
		case "enter":
			m.add = m.add.Next()

			return m.syncAddFocus(), nil
		}
	case focus == addsource.FieldAdd:
		if name == "enter" || name == "space" {
			return m.submitAdd()
		}
	case focus == addsource.FieldCancel:
		if name == "enter" || name == "space" {
			return m.closeAdd(), nil
		}
	case addsource.IsText(focus):
		if name == "enter" {
			return m.submitAdd()
		}
		in, cmd := m.addBoxes[focus].Update(key)
		m.addBoxes[focus] = in
		m.add = m.add.SetText(focus, in.Value())

		return m, cmd
	}

	return m, nil
}

// addPaste puts pasted text in the field the focus is on.
func (m Model) addPaste(text string) Model {
	focus := m.add.Focus
	if !addsource.IsText(focus) {
		return m
	}
	in := m.addBoxes[focus]
	text = strings.Join(strings.Fields(text), " ")
	in.SetValue(in.Value() + text)
	in.CursorEnd()
	m.addBoxes[focus] = in
	m.add = m.add.SetText(focus, in.Value())

	return m
}

// submitAdd adds what is written, or shows what is wrong with it.
func (m Model) submitAdd() (tea.Model, tea.Cmd) {
	result := m.add.Submit(parseLinks)
	m.add = result.State
	if !result.Done {
		return m.syncAddFocus(), nil
	}
	m = m.closeAdd()
	if !m.addable() && !m.editable() {
		m.notice = "Stop the downloads before adding to the list"

		return m, nil
	}

	return m.addAddresses(strings.Join(result.Addresses, "\n")), nil
}

// addMouse works a mouse event while the window for adding music is showing. The window covers the
// screen, so it takes every click.
func (m Model) addMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return m, nil, true
	}
	x, y := click.X-1, click.Y-1 // counted from the inside of the frame
	l := m.addLayout()
	switch y {
	case l.radioY:
		for i, span := range l.radio {
			if span.Contains(x) {
				m.add = m.add.WithMode(addsource.Mode(i)).Focused(addsource.FieldMode)

				return m.syncAddFocus(), nil, true
			}
		}
	case l.buttonY:
		switch dialog.At(l.buttons, x) {
		case 0:
			next, cmd := m.submitAdd()

			return next, cmd, true
		case 1:
			return m.closeAdd(), nil, true
		}
	}
	for f, row := range l.fieldY {
		if y == row {
			m.add = m.add.Focused(f)

			return m.syncAddFocus(), nil, true
		}
	}

	return m, nil, true
}
