package songlist

// Tone is how a status weighs on the person, which a renderer turns into a colour of its own.
type Tone int

const (
	// ToneText is a plain statement, with nothing to say about it.
	ToneText Tone = iota
	// ToneDim is background: waiting, had it, nothing to do.
	ToneDim
	// ToneFocus is something the person may want to look at.
	ToneFocus
	// ToneHighlight is something going on now.
	ToneHighlight
	// ToneSuccess is something that went well.
	ToneSuccess
	// ToneError is something that went wrong.
	ToneError
)
