package official

// isUploadOfTheSong says whether a search result is the song itself: not a cover,
// a live take, a remix or a lyric or audio-only upload of it, the title says the song
// and the title or the channel says the artist. It does not ask for the word "official" (see scoreCandidate),
// which is what tells the official video from the other uploads of the song.
func isUploadOfTheSong(res SearchResult, song Song) bool {
	if res.ID == "" || res.ID == song.OwnID || IsTopicChannel(res.Channel) || len(song.Artists) == 0 {
		return false
	}
	title, original := normalize(res.Title), normalize(song.Title)
	if hasAny(title, original, cutsToReject) || hasAny(title, original, audioMarkers) {
		return false
	}
	if TitleSimilarity(song.Title, res.Title, song.Artists) < minTitleSimilarity {
		return false
	}
	for _, artist := range song.Artists {
		if Mentions(res.Title, artist) || ChannelIsArtist(res.Channel, artist) {
			return true
		}
	}
	return false
}
