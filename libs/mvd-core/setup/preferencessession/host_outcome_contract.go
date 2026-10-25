package preferencessession

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

// RemovalOutcome is how asking to remove the app went.
type RemovalOutcome int

const (
	// RemovalStarted: the person agreed and the removal is under way. It ends with
	// the app quitting.
	RemovalStarted RemovalOutcome = iota
	// RemovalDeclined: the person said no; nothing was removed.
	RemovalDeclined
	// RemovalUnavailable: this machine has no way to ask, so nothing was removed.
	RemovalUnavailable
)
