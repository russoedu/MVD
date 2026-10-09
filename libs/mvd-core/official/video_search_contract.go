package official

// SearchResult is one video returned by a YouTube search.
type SearchResult struct {
	ID      string
	Title   string
	Channel string
	// Duration is the length in seconds, Views the view count and Verified whether
	// the channel is verified; each is zero when the search did not say.
	Duration int
	Views    int64
	Verified bool
	// MusicType is YouTube Music's tag for the video when it said (OMV, UGC, ...),
	// and Source names the database that listed it, when one did.
	MusicType string
	Source    string
}

// Searcher looks a query up on YouTube (the app wires yt-dlp here) and
// returns the matching videos in ranking order.
type Searcher func(query string) ([]SearchResult, error)
