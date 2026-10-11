package official

import (
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

// Sources are the places a song's video is looked for besides the YouTube
// search. Each is optional and may fail: a failing source is only skipped.
type Sources struct {
	// Known lists candidates a database names for the song (Wikidata), already
	// checked to exist and not to be art tracks.
	Known func(song trackpick.Song, logf func(format string, a ...interface{})) []trackpick.SearchResult
	// Music searches the videos of YouTube Music, whose results carry its tag.
	Music trackpick.Searcher
	// Type tells what an upload is (YouTube Music's tag: ATV for an art track, OMV, UGC),
	// used to drop a candidate that is an art track too: YouTube shows those under the
	// artist's own name, so a search result from "the artist's channel" may be one.
	Type func(videoID string) (string, error)
	// Identify names the song of an upload with a music database: the artist and the
	// title, or false when none says so (see package songid).
	Identify func(uploadTitle, channel string) (artist, title string, ok bool)
	// trackpick.Quality tells what a video offers to download, to choose between uploads of the
	// song when none is official.
	Quality func(videoID string) (trackpick.Quality, error)
	// Static tells whether a video does not really move: a picture with the song over
	// it, a cover, a lyric video over one background. Such an upload is not taken as a
	// better version of the song. An error means it could not tell.
	Static func(videoID string) (bool, error)
	// Cache remembers the videos found for songs.
	Cache *ResolutionCache
}
