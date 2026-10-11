// Package viewport has the arithmetic of a list in a pane: where the window of rows starts
// so the selected row stays in view, and keeping a selection inside the list.
package viewport

// Start is the index of the first of h rows that keeps row sel of n in view, the selected
// row centred while there is room on both sides.
func Start(n, sel, h int) int {
	if h <= 0 || n <= h {
		return 0
	}
	start := sel - h/2
	if start < 0 {
		start = 0
	}
	if start+h > n {
		start = n - h
	}
	return start
}

// Clamp keeps v between lo and hi; when hi is below lo the result is lo.
func Clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
