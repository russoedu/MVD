package songfile

// Parse reads the songs of a text: a CSV whose first row names a title and an
// artist column is read by its columns, anything else as "Artist - Title" lines.
func Parse(text string) List {
	if songs, skipped, ok := ParseCSV(text); ok && len(songs) > 0 {
		return List{Songs: songs, Skipped: skipped}
	}
	songs, skipped := ParseLines(text)
	return List{Songs: songs, Skipped: skipped}
}

// IsCSV reports whether Parse reads the text by columns.
func IsCSV(text string) bool {
	songs, _, ok := ParseCSV(text)
	return ok && len(songs) > 0
}
