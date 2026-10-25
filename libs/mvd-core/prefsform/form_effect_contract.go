package prefsform

// Effect is something the form asks its host to do. The host does it and answers with Picked or
// ActionDone.
type Effect interface{ isEffect() }

// PickFolder asks for the operating system's folder chooser, starting at Start, for the setting
// Field of the page.
type PickFolder struct {
	Field int
	Start string
}

// RunAction asks for one of the host's actions: moving the app, or uninstalling it.
type RunAction struct{ Kind ActionKind }

func (PickFolder) isEffect() {}
func (RunAction) isEffect()  {}
