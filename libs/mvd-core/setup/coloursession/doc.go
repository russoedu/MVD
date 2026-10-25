// Package coloursession is renderer-neutral: it holds what the colours screen does. Which colour the
// cursor is on, the colour being typed, whether it is valid, the default a colour goes back to, and
// the words the screen says. A renderer draws it, owns the text box, and applies the colours to its
// own drawing when the session says so (see Effect).
package coloursession
