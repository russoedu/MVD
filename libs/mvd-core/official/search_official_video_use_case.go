package official

import (
	"fmt"
	"youtube-downloader/libs/mvd-core/official/titlematch"
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

// songOfUpload is the song an upload stands for, as it is searched for: the one a
// music database names when it knows ("Tonight Is The Night" by "Le Click" for an
// upload titled "Tonight Is The Night (Le Click - Dance Mix)"), else the upload's
// own title and the artist its channel suggests. named is false when the database
// did not know and the upload is a plain video, whose title and channel say little
// about the song.
func (r *Resolver) songOfUpload(videoID, title, channel string, durationSec int, artTrack bool, logf func(format string, a ...interface{})) (song trackpick.Song, named bool) {
	song = trackpick.Song{Title: title, Artists: artistNames(titlematch.ArtistFromChannel(channel)), DurationSec: durationSec, OwnID: videoID, OwnIsStill: artTrack}
	// The Music card of the watch page names the song exactly, whatever the upload's own
	// title puts around it ("As 7 Melhores Volume 1- Bill Power - I Swear"): when it is
	// there no other source is asked.
	if !artTrack {
		if page, err := r.get(r.WatchBase + videoID); err == nil {
			if artist, name, ok := SongFromWatchPage(page); ok {
				logf("[official] %s: the music card of the video names it %q by %q", videoID, name, artist)
				song.Title, song.Artists = name, artistNames(artist)
				return song, true
			}
		}
	}
	if r.Sources.Identify != nil && title != "" {
		if artist, name, ok := r.Sources.Identify(title, channel); ok {
			logf("[official] %s: a music database names it %q by %q", videoID, name, artist)
			song.Title, song.Artists = name, artistNames(artist)
			return song, true
		}
	}
	return song, artTrack
}

// fromSearch looks the song up on YouTube by its title and artist and returns the id
// of the best video, and why, or "" when the search finds nothing convincing. It also
// returns every result the searches saw, for choosing among them when none is official.
func (r *Resolver) fromSearch(song trackpick.Song, logf func(format string, a ...interface{})) (string, string, []trackpick.SearchResult) {
	if r.Searcher == nil || song.Title == "" {
		return "", "", nil
	}

	pick, ok, results, err := FindBestVideo(r.Searcher, r.Sources, song, func(format string, a ...interface{}) {
		logf("[official] %s: "+format, append([]interface{}{song.OwnID}, a...)...)
	})
	if err != nil {
		logf("[official] %s: %v", song.OwnID, err)
		return "", "", nil
	}
	if !ok {
		return "", "", results
	}
	return pick.ID, fmt.Sprintf("found by searching: %s by %q (%s)", pick.Kind, pick.Channel, pick.Why), results
}
