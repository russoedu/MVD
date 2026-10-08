package official

// SearchResult is one video returned by a YouTube search.
type SearchResult struct {
	ID      string
	Title   string
	Channel string
}

// Searcher looks a query up on YouTube (the app wires yt-dlp here) and
// returns the matching videos in ranking order.
type Searcher func(query string) ([]SearchResult, error)
