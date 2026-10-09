package official

// Sources are the places a song's video is looked for besides the YouTube
// search. Each is optional and may fail: a failing source is only skipped.
type Sources struct {
	// Known lists candidates a database names for the song (Wikidata), already
	// checked to exist and not to be art tracks.
	Known func(song Song, logf func(format string, a ...interface{})) []SearchResult
	// Music searches the videos of YouTube Music, whose results carry its tag.
	Music Searcher
	// Type tells what an upload is (YouTube Music's tag: ATV for an art track, OMV, UGC),
	// used to drop a candidate that is an art track too: YouTube shows those under the
	// artist's own name, so a search result from "the artist's channel" may be one.
	Type func(videoID string) (string, error)
	// Cache remembers the videos found for songs.
	Cache *ResolutionCache
}

// maxKnownCandidates is how many of the videos a database lists for a song are
// checked: each costs a request.
const maxKnownCandidates = 4

// KnownCandidates returns a Sources.Known that asks Wikidata for the videos of
// a song and keeps those that exist, are not art tracks (YouTube Music's tag,
// when it answers) and are not the song's own upload. Wikidata lists many a
// song's art track as its video ("Lovely Day"), which would be no improvement.
func (r *Resolver) KnownCandidates(wikidata *WikidataClient) func(Song, func(string, ...interface{})) []SearchResult {
	return func(song Song, logf func(format string, a ...interface{})) []SearchResult {
		ids, err := wikidata.VideoIDs(song.Title, song.Artists)
		if err != nil {
			logf("Wikidata could not be asked: %v", err)
			return nil
		}

		var candidates []SearchResult
		for _, id := range ids {
			if len(candidates) == maxKnownCandidates {
				break
			}
			if id == song.OwnID {
				continue
			}
			author, title, ok, err := r.lookupVideo(id)
			if err != nil || !ok || IsTopicChannel(author) {
				continue
			}

			musicType := ""
			if r.TrackInfos != nil {
				if info, err := r.TrackInfos(id); err == nil {
					musicType = info.Type
				}
			}
			if artTrack, _ := ArtTrackFromVideoType(musicType); artTrack {
				logf("Wikidata lists %s, which is an art track too", id)
				continue
			}
			candidates = append(candidates, SearchResult{ID: id, Title: title, Channel: author, MusicType: musicType, Source: "Wikidata"})
		}
		return candidates
	}
}
