package songfile

// Song is one song of a list.
type Song struct {
	Title string
	// Artist is the artists, comma separated.
	Artist     string
	DurationMs int
}

// List is a song file and its songs, in order.
type List struct {
	Title string
	Songs []Song
	// Skipped counts the lines that could not be read as a song.
	Skipped int
}
