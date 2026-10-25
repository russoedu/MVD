// Package preferencessession is renderer-neutral: it holds what the preferences screen does. Which
// setting the cursor is on, how each is changed (a switch, a number, a text, a choice, a folder),
// what is accepted, what each setting reads as, and the questions the host is asked to put to the
// person (choose a folder, remove the app, move the app). The host runs what the session asks of it
// through an Effect and gives the answer back. A renderer draws the session, owns the text box and
// the folder browser, and is the one that knows which key does what.
package preferencessession
