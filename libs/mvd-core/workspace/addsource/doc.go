// Package addsource is renderer-neutral: it is the window for adding music to a list, apart from how
// it is drawn. The person chooses between a platform address and a search by name; what is written,
// where the focus is, what each choice explains and what a written entry becomes (the addresses that
// go in the list) are its business. The text boxes belong to the renderer, which pushes what is typed
// up with SetText.
package addsource
