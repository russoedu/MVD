package download

// thumb is where the bar of a scrollbar sits on a track of h cells for n rows scrolled down by top, and
// how long it is. ok is false when every row shows, and there is nothing to scroll.
func thumb(n, h, top int) (start, length int, ok bool) {
	if h <= 0 || n <= h {
		return 0, 0, false
	}
	top = min(max(top, 0), n-h)
	length = max(1, h*h/n)
	start = top * (h - length) / (n - h)

	return start, length, true
}

// topAt is how far down a pane with n rows, h of them showing, is scrolled when the bar is taken to
// cell at of a track h cells long: the middle of the bar goes there.
func topAt(n, h, at int) int {
	if h <= 0 || n <= h {
		return 0
	}
	_, length, _ := thumb(n, h, 0)
	span := h - length
	if span <= 0 {
		return 0
	}
	start := min(max(at-length/2, 0), span)

	return (start*(n-h) + span/2) / span
}
