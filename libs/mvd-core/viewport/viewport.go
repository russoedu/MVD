// Package viewport keeps a selected row in view: windowing a list into the room a pane has,
// and the arithmetic that lets a click be matched to the row that was drawn there.
package viewport

// Rows returns the slice of rows that keeps index sel visible in h rows.
func Rows[T any](rows []T, sel, h int) []T {
	if h <= 0 || len(rows) <= h {
		return rows
	}
	start := Start(len(rows), sel, h)
	return rows[start : start+h]
}

// Start is the index of the first row Rows shows: the same arithmetic, so a click
// can be matched to the row that was drawn there.
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
