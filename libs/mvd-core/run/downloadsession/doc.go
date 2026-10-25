// Package downloadsession is renderer-neutral: it is what the download screen does, with no
// terminal in it. A Session holds what the person has chosen and is looking at (the pane they are
// in, the playlist and song they are on, which songs are ticked, how the songs are filtered and
// sorted, which columns show, a notice, a question waiting for an answer) together with what the
// run has reported. Its methods are the intents: what the person means, never a key or a position
// on a screen. A method that needs the host to do something returns an Effect.
//
// A renderer draws a Session and turns what the person does into calls on it. Sizes in cells or
// pixels, scroll offsets, hit-testing, colours and the widgets that edit text belong to the
// renderer.
package downloadsession
