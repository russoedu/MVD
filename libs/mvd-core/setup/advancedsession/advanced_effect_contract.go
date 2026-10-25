package advancedsession

// Effect is what the host has to do after an intent.
type Effect int

const (
	// NoEffect needs nothing from the host.
	NoEffect Effect = iota
	// Save asks the host to keep the settings and leave the screen.
	Save
	// Cancel asks the host to leave the screen.
	Cancel
)
