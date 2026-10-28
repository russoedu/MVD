package runcontrol

// RunMenu says which entries of the Run menu, and which buttons of the toolbar, can be used now.
type RunMenu struct {
	Start, Pause, Resume, Stop, Retry, DownloadAll bool
}

// RunMenuOf decides the Run menu from what the run is doing. nothing says the list has no
// addresses, and toolsWaiting that the tools of the downloads are not ready.
func RunMenuOf(f Facts, nothing, toolsWaiting bool) RunMenu {
	busy := f.Busy()
	running := f.Active()

	return RunMenu{
		Start:       !running && !busy && !nothing && !toolsWaiting,
		Pause:       running && !busy && !f.Paused,
		Resume:      running && !busy && f.Paused,
		Stop:        running && !busy,
		Retry:       f.HasRun && !busy && f.Tally.Failed > 0,
		DownloadAll: running && !busy && f.Waiting(),
	}
}

// DisabledReason is why a button of the toolbar cannot be used now: the run is changing, the tools
// are being fetched (toolsMessage says how far), or the list is empty.
func DisabledReason(f Facts, toolsWaiting bool, toolsMessage string) string {
	switch {
	case f.Busy():
		return "One moment..."
	case toolsWaiting:
		return toolsMessage
	}

	return "Add a source first."
}

// Refusal is why the downloads cannot begin now, as the notice the person is told; "" when nothing
// stands in the way. sources is how many addresses the list has, and hasStart says the host can start
// downloads at all.
func Refusal(f Facts, sources int, hasStart bool) string {
	switch {
	case f.HasRun:
		return "The downloads are already going"
	case sources == 0:
		return "The list is empty: add a source first (Ctrl+L)"
	case !hasStart:
		return "Downloading is not available here"
	}

	return ""
}
