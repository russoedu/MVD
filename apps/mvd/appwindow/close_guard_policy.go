package appwindow

// keepOpen decides whether closing the window is stopped: when downloads are running the
// person is asked, and the window stays unless they agree to quit.
func keepOpen(busy func() bool, askQuit func() bool) bool {
	if busy == nil || !busy() {
		return false
	}
	return !askQuit()
}
