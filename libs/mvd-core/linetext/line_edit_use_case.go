package linetext

// Paste puts text where the cursor is, in place of what is selected. Line breaks become
// spaces: it is one line.
func (t Text) Paste(text string) Text {
	clean := make([]rune, 0, len(text))
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		clean = append(clean, r)
	}

	return t.Insert(string(clean))
}

// Insert puts text where the cursor is, in place of what is selected, as it is.
func (t Text) Insert(text string) Text {
	add := []rune(text)
	lo, hi := t.Bounds()
	out := make([]rune, 0, len(t.text)-(hi-lo)+len(add))
	out = append(out, t.text[:lo]...)
	out = append(out, add...)
	t.text = append(out, t.text[hi:]...)
	t.caret = lo + len(add)
	t.anchor = t.caret

	return t
}

// Delete takes out what is selected, or the character before the cursor (back) or at it.
func (t Text) Delete(back bool) Text {
	if !t.HasSelection() {
		if back && t.caret > 0 {
			t.anchor = t.caret - 1
		} else if !back && t.caret < len(t.text) {
			t.anchor = t.caret + 1
		}
	}

	return t.Insert("")
}

// MoveTo puts the cursor at to; with extend the selection grows to it, else it ends.
func (t Text) MoveTo(to int, extend bool) Text {
	t.caret = max(min(to, len(t.text)), 0)
	if !extend {
		t.anchor = t.caret
	}

	return t
}

// Step moves the cursor one character left (-1) or right (1). With extend the selection grows;
// without it, a selection ends at its near edge, as in any text field.
func (t Text) Step(dir int, extend bool) Text {
	if t.HasSelection() && !extend {
		lo, hi := t.Bounds()
		if dir < 0 {
			return t.MoveTo(lo, false)
		}

		return t.MoveTo(hi, false)
	}

	return t.MoveTo(t.caret+dir, extend)
}

// SelectAll selects the whole text.
func (t Text) SelectAll() Text {
	t.anchor, t.caret = 0, len(t.text)

	return t
}
