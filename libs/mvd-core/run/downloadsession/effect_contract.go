package downloadsession

// Effect is something the host has to do after an intent, because the session cannot: leave the
// screen, show a page in the person's browser. A renderer turns each into whatever its host uses.
type Effect interface{ isEffect() }

// Quit asks the host to leave the screen.
type Quit struct{}

// OpenURL asks the host to show a page in the person's browser.
type OpenURL struct{ URL string }

// RunEnded tells the host that the run behind a hosted screen has ended: its events are over and
// the screen keeps showing what it reported.
type RunEnded struct{}

func (Quit) isEffect()     {}
func (OpenURL) isEffect()  {}
func (RunEnded) isEffect() {}
