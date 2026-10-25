package coloursession

// Effect is what the host has to do after an intent.
type Effect int

const (
	// NoEffect needs nothing from the host.
	NoEffect Effect = iota
	// ApplyColours asks the host to draw with the colours of the session from now on, so the screen
	// itself previews a choice.
	ApplyColours
	// Save asks the host to keep the colours and leave the screen.
	Save
	// Cancel asks the host to leave the screen.
	Cancel
)
