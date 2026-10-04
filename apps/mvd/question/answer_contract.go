package question

// Answer is what the person chose in the question about moving the app.
type Answer int

const (
	// AnswerLeave is No, Cancel, or closing the box.
	AnswerLeave Answer = iota
	// AnswerFirst is the first choice offered (everyone, or the only move there is).
	AnswerFirst
	// AnswerSecond is the second choice offered (just for me).
	AnswerSecond
	// AnswerUnavailable means there was no way to ask, so nothing was asked.
	AnswerUnavailable
)
