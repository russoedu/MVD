package official

import "fmt"

// fromSearch looks the track up on YouTube by its title and artist and returns
// the id of the best video, and why, or "" when the search finds nothing
// convincing. durationSec is the length of the art track, or 0 when unknown.
func (r *Resolver) fromSearch(videoID, title, channel string, durationSec int, logf func(format string, a ...interface{})) (string, string) {
	if r.Searcher == nil || title == "" {
		return "", ""
	}

	song := Song{
		Title:       title,
		Artists:     artistNames(ArtistFromChannel(channel)),
		DurationSec: durationSec,
		OwnID:       videoID,
	}
	pick, ok, _, err := FindBestVideo(r.Searcher, song, func(format string, a ...interface{}) {
		logf("[official] %s: "+format, append([]interface{}{videoID}, a...)...)
	})
	if err != nil {
		logf("[official] %s: %v", videoID, err)
		return "", ""
	}
	if !ok {
		return "", ""
	}
	return pick.ID, fmt.Sprintf("found by searching: %s by %q (%s)", pick.Kind, pick.Channel, pick.Why)
}
