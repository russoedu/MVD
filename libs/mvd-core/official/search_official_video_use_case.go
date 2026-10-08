package official

// fromSearch looks the track up on YouTube as "<title> <artist> official
// video" and returns the id of the official upload, or "" when the search
// finds nothing convincing.
func (r *Resolver) fromSearch(videoID, title, channel string, logf func(format string, a ...interface{})) string {
	if r.Searcher == nil || title == "" {
		return ""
	}
	artist := ArtistFromChannel(channel)
	query := SearchQuery(title, artist)
	results, err := r.Searcher(query)
	if err != nil {
		logf("[official] %s: search for %q failed: %v", videoID, query, err)
		return ""
	}
	found := PickSearchResult(results, title, artist, videoID)
	if found == "" {
		logf("[official] %s: search for %q found no official video", videoID, query)
		return ""
	}
	logf("[official] %s: search for %q found %s", videoID, query, found)
	return found
}
