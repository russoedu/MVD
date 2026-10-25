// Package songpaste is renderer-neutral: it holds what the screen for pasting songs does, apart from
// how it is drawn. The text pasted, how it reads (how many songs, how many lines no song could be
// read from), and saving it as a file for the download list to name are its business. The text box
// belongs to the renderer, which pushes what is typed up with SetText.
package songpaste
