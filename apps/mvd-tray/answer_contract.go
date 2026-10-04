package main

// answer is what the person chose in the question about moving the app.
type answer int

const (
	// answerLeave is No, Cancel, or closing the box.
	answerLeave answer = iota
	// answerFirst is the first choice offered (everyone, or the only move there is).
	answerFirst
	// answerSecond is the second choice offered (just for me).
	answerSecond
	// answerUnavailable means there was no way to ask, so nothing was asked.
	answerUnavailable
)
