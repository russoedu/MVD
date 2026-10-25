package closequestion

import "fmt"

// The answers, in the order they are offered.
const (
	AnswerQuit = iota
	AnswerKeep
)

// Answers are what the two answers are called.
var Answers = []string{"Quit", "Keep downloading"}

// Outcome is what the person decided.
type Outcome int

// The outcomes.
const (
	// Undecided: the question stays.
	Undecided Outcome = iota
	// QuitNow: stop the downloads and close the app.
	QuitNow
	// Dismissed: the question goes and the downloads go on.
	Dismissed
)

// State is the question: whether it is asked, and which answer has the focus. It is a value.
type State struct {
	On    bool
	Focus int
}

// Ask is the question asked, with the safe answer in focus.
func Ask() State { return State{On: true, Focus: AnswerKeep} }

// Words are the two lines of the question for n songs that would be lost.
func Words(n int) (what, ask string) {
	what = fmt.Sprintf("%d songs are still downloading.", n)
	if n == 1 {
		what = "1 song is still downloading."
	}

	return what, "Quit and stop them?"
}

// Key is what a key does while the question is asked: y quits, n and esc keep downloading, the
// arrows and tab move the focus, and enter or space take the answer in focus.
func (s State) Key(key string) (State, Outcome) {
	switch key {
	case "y", "Y":
		return State{}, QuitNow
	case "n", "N", "esc":
		return State{}, Dismissed
	case "left", "right", "tab", "shift+tab", "h", "l":
		s.Focus = 1 - s.Focus
	case "enter", "space":
		if s.Focus == AnswerQuit {
			return State{}, QuitNow
		}

		return State{}, Dismissed
	}

	return s, Undecided
}

// Hover gives an answer the focus, as the pointer over it does.
func (s State) Hover(answer int) State {
	s.Focus = answer

	return s
}

// Click is a click on an answer.
func (s State) Click(answer int) (State, Outcome) {
	if answer == AnswerQuit {
		return State{}, QuitNow
	}

	return State{}, Dismissed
}
