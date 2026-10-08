package spotify

// Track is one song of a playlist.
type Track struct {
	Title string
	// Artist is the artists as Spotify shows them, comma separated.
	Artist     string
	DurationMs int
}

// Playlist is a public playlist and its tracks, in order.
type Playlist struct {
	Title  string
	Tracks []Track
}
