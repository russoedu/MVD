package urllist

// Outcome is where the screen leads when the person asks for it: the host acts on it.
type Outcome int

const (
	// None stays on the screen.
	None Outcome = iota
	// Start begins the downloads.
	Start
	// Prefs opens the preferences.
	Prefs
	// Songs opens the screen for pasting songs.
	Songs
	// Review opens the playlist editor on the links.
	Review
	// Quit leaves the program.
	Quit
)
