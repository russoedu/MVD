package prefspopup

// scrollbarThumb says, for each of the rows lines a page of total lines shows, scrolled to
// scroll, whether the thumb of a scrollbar stands on it. Nothing does when the whole page fits.
// The thumb is as tall as the share of the page that shows, at least one line, and sits where
// the window is among the lines.
func scrollbarThumb(total, rows, scroll int) []bool {
	thumb := make([]bool, max(rows, 0))
	if rows <= 0 || total <= rows {
		return thumb
	}
	size := max(rows*rows/total, 1)
	top := 0
	if hidden := total - rows; hidden > 0 {
		top = min(scroll, hidden) * (rows - size) / hidden
	}
	for i := top; i < top+size && i < rows; i++ {
		thumb[i] = true
	}
	return thumb
}
