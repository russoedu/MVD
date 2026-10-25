// Package lineedit is a one-line text editor for a row of the screen: the cursor and a
// selection, both set with the keyboard or the mouse, and copy, cut and paste. It draws
// itself in the room it is given and scrolls to keep the cursor in view. The text and what
// editing does to it are in package linetext; this is the terminal's side of it.
package lineedit

import (
	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/linetext"
)

// Model is the text being edited.
type Model struct {
	text linetext.Text
	// scroll is the first character that shows, and width the cells the editor has.
	scroll, width int
	// selecting is true while the pointer is held down to select with it.
	selecting bool
}

// New returns an editor for text, with all of it selected so that typing replaces it.
func New(text string, width int) Model {
	m := Model{text: linetext.New(text), width: max(width, 1)}

	return m.reveal()
}

// Value is the text.
func (m Model) Value() string { return m.text.Value() }

// Selection is the text selected, "" for none.
func (m Model) Selection() string { return m.text.Selection() }

// HasSelection reports whether some text is selected.
func (m Model) HasSelection() bool { return m.text.HasSelection() }

// SetWidth gives the editor the cells it draws in.
func (m Model) SetWidth(width int) Model {
	m.width = max(width, 1)

	return m.reveal()
}

// Paste puts text where the cursor is, in place of what is selected. Line breaks become
// spaces: it is one line.
func (m Model) Paste(text string) Model { return m.edited(m.text.Paste(text)) }

// edited is the editor with the text after an edit, scrolled so the cursor shows.
func (m Model) edited(t linetext.Text) Model {
	m.text = t

	return m.reveal()
}

// reveal scrolls so the cursor shows.
func (m Model) reveal() Model {
	caret := m.text.Caret()
	if caret < m.scroll {
		m.scroll = caret
	}
	if caret >= m.scroll+m.width {
		m.scroll = caret - m.width + 1
	}
	m.scroll = max(min(m.scroll, m.text.Len()), 0)

	return m
}

// Update works a key or the mouse. A mouse position is in the cells of the editor, from its
// left edge. Copy and cut put the selection on the clipboard; paste asks for it, and the
// answer (a tea.ClipboardMsg) is for Paste.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.key(msg)
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}
		m = m.edited(m.text.MoveTo(m.scroll+msg.X, false))
		m.selecting = true

		return m, nil
	case tea.MouseMotionMsg:
		if m.selecting {
			m = m.edited(m.text.MoveTo(m.scroll+msg.X, true))
		}

		return m, nil
	case tea.MouseReleaseMsg:
		m.selecting = false

		return m, nil
	}

	return m, nil
}

// SelectAll selects the whole text.
func (m Model) SelectAll() Model { return m.edited(m.text.SelectAll()) }

// Copy puts the selection on the clipboard.
func (m Model) Copy() tea.Cmd {
	if !m.HasSelection() {
		return nil
	}

	return tea.SetClipboard(m.Selection())
}

// Cut copies the selection and takes it out.
func (m Model) Cut() (Model, tea.Cmd) {
	return m.edited(m.text.Delete(false)), m.Copy()
}

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	shift := k.Mod.Contains(tea.ModShift)
	switch k.String() {
	case "left", "shift+left":
		return m.edited(m.text.Step(-1, shift)), nil
	case "right", "shift+right":
		return m.edited(m.text.Step(1, shift)), nil
	case "home", "shift+home":
		return m.edited(m.text.MoveTo(0, shift)), nil
	case "end", "shift+end":
		return m.edited(m.text.MoveTo(m.text.Len(), shift)), nil
	case "backspace":
		return m.edited(m.text.Delete(true)), nil
	case "delete":
		return m.edited(m.text.Delete(false)), nil
	case "ctrl+a":
		return m.SelectAll(), nil
	case "ctrl+c":
		return m, m.Copy()
	case "ctrl+x":
		return m.Cut()
	case "ctrl+v":
		return m, tea.ReadClipboard
	}
	if k.Text != "" && !k.Mod.Contains(tea.ModCtrl) && !k.Mod.Contains(tea.ModAlt) {
		return m.edited(m.text.Insert(k.Text)), nil
	}

	return m, nil
}
