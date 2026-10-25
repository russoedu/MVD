// Package linetext is renderer-neutral: the text of a one-line editor with a cursor and a
// selection, and what editing does to them. Where the cursor shows, how wide the line is and which
// key does what are the renderer's.
package linetext

// Text is the text being edited. It is a value: every change gives a new Text.
type Text struct {
	text []rune
	// caret is where typing goes; anchor is the other end of the selection, equal to caret
	// when nothing is selected.
	caret, anchor int
}

// New returns text with all of it selected, so that typing replaces it.
func New(text string) Text {
	t := Text{text: []rune(text)}
	t.anchor, t.caret = 0, len(t.text)

	return t
}

// Value is the text.
func (t Text) Value() string { return string(t.text) }

// Len is how many characters the text has.
func (t Text) Len() int { return len(t.text) }

// Rune is the character at i, which must be inside the text.
func (t Text) Rune(i int) rune { return t.text[i] }

// Caret is where typing goes, in characters from the start.
func (t Text) Caret() int { return t.caret }

// Selection is the text selected, "" for none.
func (t Text) Selection() string {
	lo, hi := t.Bounds()

	return string(t.text[lo:hi])
}

// HasSelection reports whether some text is selected.
func (t Text) HasSelection() bool { return t.caret != t.anchor }

// Bounds are the start and the end of the selection; equal when nothing is selected.
func (t Text) Bounds() (lo, hi int) {
	return min(t.caret, t.anchor), max(t.caret, t.anchor)
}
