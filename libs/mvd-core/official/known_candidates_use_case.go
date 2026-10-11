package official

import (
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

// maxKnownCandidates is how many of the videos a database lists for a song are
// checked: each costs a request.
const maxKnownCandidates = 4

// KnownCandidates returns a Sources.Known that asks Wikidata for the videos of
// a song and keeps those that exist, are not art tracks (YouTube Music's tag,
// when it answers) and are not the song's own upload. Wikidata lists many a
// song's art track as its video ("Lovely Day"), which would be no improvement.
func (r *Resolver) KnownCandidates(wikidata *WikidataClient) func(trackpick.Song, func(string, ...interface{})) []trackpick.SearchResult {
	return func(song trackpick.Song, logf func(format string, a ...interface{})) []trackpick.SearchResult {
		ids, err := wikidata.VideoIDs(song.Title, song.Artists)
		if err != nil {
			logf("Wikidata could not be asked: %v", err)
			return nil
		}

		var candidates []trackpick.SearchResult
		for _, id := range ids {
			if len(candidates) == maxKnownCandidates {
				break
			}
			if id == song.OwnID {
				continue
			}
			author, title, ok, err := r.lookupVideo(id)
			if err != nil || !ok || trackpick.IsTopicChannel(author) {
				continue
			}

			musicType := ""
			if r.TrackInfos != nil {
				if info, err := r.TrackInfos(id); err == nil {
					musicType = info.Type
				}
			}
			if artTrack, _ := trackpick.ArtTrackFromVideoType(musicType); artTrack {
				logf("Wikidata lists %s, which is an art track too", id)
				continue
			}
			candidates = append(candidates, trackpick.SearchResult{ID: id, Title: title, Channel: author, MusicType: musicType, Source: "Wikidata"})
		}
		return candidates
	}
}
