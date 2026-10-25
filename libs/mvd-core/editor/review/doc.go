// Package review is renderer-neutral: it is what the playlist editor does with a plan before
// anything is downloaded. The songs of a plan, what was proposed for each, what the person
// decides about them, the questions asked on the way, the saving and opening of a review file, and
// the words the screen says of all that. It knows nothing of keys, cells or colours: a renderer
// turns what the person does into calls on a Session, runs the Effects they return, and draws
// the state.
package review
