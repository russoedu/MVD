package tui

// MoveOutcome is how asking to move the app to its own folder went.
type MoveOutcome int

const (
	// MoveStarted: the person agreed and the app was moved and started from its new
	// place. It ends with this copy quitting.
	MoveStarted MoveOutcome = iota
	// MoveDeclined: the person said no, or the move failed and said so.
	MoveDeclined
	// MoveAlreadyThere: the app already lives in its own folder.
	MoveAlreadyThere
	// MoveUnavailable: there is nowhere to move it to on this machine, or no way to ask.
	MoveUnavailable
)

// Mover asks the person, on the machine's own screen, whether to move the app to its
// own folder, and moves it if they agree. It waits for the answers, which may take as
// long as the person does, and works whatever the person answered when it asked by
// itself. A host that can move the app passes one to the setup screens, which then offer
// it on the preferences.
type Mover func() (MoveOutcome, error)

// moveAnsweredMsg is the answer of a Mover.
type moveAnsweredMsg struct {
	outcome MoveOutcome
	err     error
}
