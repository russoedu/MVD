// Package linequestion is what the one-line question at the bottom of the workspace decides, with
// no terminal in it: whether a key ends the question, what the answer is, and what the question
// says when a link has to be asked for again.
package linequestion

import "strings"

// Outcome is what a key does to the question.
type Outcome int

const (
	// Typing leaves the question open: the key edits the line.
	Typing Outcome = iota
	// Cancelled closes the question with no answer.
	Cancelled
	// Empty closes the question: it was answered with nothing.
	Empty
	// Answered closes the question with a text.
	Answered
)

// Decide says what key does to a question whose line holds value, and the answer when there is one.
func Decide(key, value string) (Outcome, string) {
	switch key {
	case "esc":
		return Cancelled, ""
	case "enter":
		text := strings.TrimSpace(value)
		if text == "" {
			return Empty, ""
		}

		return Answered, text
	}

	return Typing, ""
}

// LinkPrompt is what the question says when it asks for a link, or for the one being changed.
func LinkPrompt(changing bool) string {
	if changing {
		return "Change the link"
	}

	return "Add a playlist or video link"
}
