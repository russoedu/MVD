// Package lineedit is a one-line text editor for a row of the screen: the cursor and a
// selection, both set with the keyboard or the mouse, and copy, cut and paste. It draws
// itself in the room it is given and scrolls to keep the cursor in view.
package lineedit

import (
	tea "charm.land/bubbletea/v2"
)

// Model is the text being edited.
type Model struct {
	text []rune
	// caret is where typing goes; anchor is the other end of the selection, equal to caret
	// when nothing is selected.
	caret, anchor int
	// scroll is the first character that shows, and width the cells the editor has.
	scroll, width int
	// selecting is true while the pointer is held down to select with it.
	selecting bool
}

// New returns an editor for text, with all of it selected so that typing replaces it.
func New(text string, width int) Model {
	m := Model{text: []rune(text), width: max(width, 1)}
	m.anchor, m.caret = 0, len(m.text)
	return m.reveal()
}

// Value is the text.
func (m Model) Value() string { return string(m.text) }

// Selection is the text selected, "" for none.
func (m Model) Selection() string {
	lo, hi := m.bounds()
	return string(m.text[lo:hi])
}

// HasSelection reports whether some text is selected.
func (m Model) HasSelection() bool { return m.caret != m.anchor }

// SetWidth gives the editor the cells it draws in.
func (m Model) SetWidth(width int) Model {
	m.width = max(width, 1)
	return m.reveal()
}

func (m Model) bounds() (lo, hi int) {
	return min(m.caret, m.anchor), max(m.caret, m.anchor)
}

// Paste puts text where the cursor is, in place of what is selected. Line breaks become
// spaces: it is one line.
func (m Model) Paste(text string) Model {
	clean := make([]rune, 0, len(text))
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		clean = append(clean, r)
	}
	return m.insert(clean)
}

func (m Model) insert(add []rune) Model {
	lo, hi := m.bounds()
	out := make([]rune, 0, len(m.text)-(hi-lo)+len(add))
	out = append(out, m.text[:lo]...)
	out = append(out, add...)
	m.text = append(out, m.text[hi:]...)
	m.caret = lo + len(add)
	m.anchor = m.caret
	return m.reveal()
}

// remove takes out what is selected, or the character before the cursor (back) or at it.
func (m Model) remove(back bool) Model {
	if !m.HasSelection() {
		if back && m.caret > 0 {
			m.anchor = m.caret - 1
		} else if !back && m.caret < len(m.text) {
			m.anchor = m.caret + 1
		}
	}
	return m.insert(nil)
}

// move puts the cursor at to; with extend the selection grows to it, else it ends.
func (m Model) move(to int, extend bool) Model {
	m.caret = max(min(to, len(m.text)), 0)
	if !extend {
		m.anchor = m.caret
	}
	return m.reveal()
}

// reveal scrolls so the cursor shows.
func (m Model) reveal() Model {
	if m.caret < m.scroll {
		m.scroll = m.caret
	}
	if m.caret >= m.scroll+m.width {
		m.scroll = m.caret - m.width + 1
	}
	m.scroll = max(min(m.scroll, len(m.text)), 0)
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
		m = m.move(m.scroll+msg.X, false)
		m.selecting = true
		return m, nil
	case tea.MouseMotionMsg:
		if m.selecting {
			m = m.move(m.scroll+msg.X, true)
		}
		return m, nil
	case tea.MouseReleaseMsg:
		m.selecting = false
		return m, nil
	}
	return m, nil
}

// SelectAll selects the whole text.
func (m Model) SelectAll() Model {
	m.anchor, m.caret = 0, len(m.text)
	return m.reveal()
}

// Copy puts the selection on the clipboard.
func (m Model) Copy() tea.Cmd {
	if !m.HasSelection() {
		return nil
	}
	return tea.SetClipboard(m.Selection())
}

// Cut copies the selection and takes it out.
func (m Model) Cut() (Model, tea.Cmd) {
	return m.remove(false), m.Copy()
}

func (m Model) key(k tea.KeyPressMsg) (Model, tea.Cmd) {
	shift := k.Mod.Contains(tea.ModShift)
	switch k.String() {
	case "left", "shift+left":
		if m.HasSelection() && !shift {
			lo, _ := m.bounds()
			return m.move(lo, false), nil
		}
		return m.move(m.caret-1, shift), nil
	case "right", "shift+right":
		if m.HasSelection() && !shift {
			_, hi := m.bounds()
			return m.move(hi, false), nil
		}
		return m.move(m.caret+1, shift), nil
	case "home", "shift+home":
		return m.move(0, shift), nil
	case "end", "shift+end":
		return m.move(len(m.text), shift), nil
	case "backspace":
		return m.remove(true), nil
	case "delete":
		return m.remove(false), nil
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
		return m.insert([]rune(k.Text)), nil
	}
	return m, nil
}
