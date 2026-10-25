// Package urllist is renderer-neutral: it holds what the download-list screen does, apart from how
// it is drawn. The text of the list, the question whether to discard it, which lines are links, and
// what each intent (start, review, preferences, songs, reset, leave) leads to are its business. The
// text box belongs to the renderer, which pushes what is typed up with SetText.
package urllist
