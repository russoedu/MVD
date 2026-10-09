package official

import "fmt"

// VideoTyper returns YouTube Music's tag for an upload ("ATV", "OMV", "UGC",
// ...), or "" when it has none. The app wires YouTubeMusicClient.VideoType
// here; it is optional.
type VideoTyper func(videoID string) (string, error)

// isArtTrack decides whether an upload is an auto-generated art track, the only
// kind worth looking an official video up for. Many art tracks show the
// artist's own name as the channel instead of "<Artist> - Topic", so the name
// says nothing: YouTube Music's tag is asked first, then the channel name, then
// the description on the watch page. The returned reason says why an upload was
// left alone.
//
// Without a VideoTyper every upload is treated as an art track, as before.
func (r *Resolver) isArtTrack(videoID, channel string, logf func(format string, a ...interface{})) (bool, string) {
	if r.VideoTypes == nil {
		return true, ""
	}

	videoType, err := r.VideoTypes(videoID)
	if err != nil {
		logf("[official] %s: YouTube Music could not say what this upload is (%v)", videoID, err)
	}
	if artTrack, known := ArtTrackFromVideoType(videoType); known {
		if artTrack {
			logf("[official] %s: an auto-generated track (YouTube Music: %s)", videoID, videoType)
			return true, ""
		}
		return false, fmt.Sprintf("already a video, not an art track (YouTube Music: %s)", videoType)
	}

	if IsTopicChannel(channel) {
		return true, ""
	}
	page, err := r.get(r.WatchBase + videoID)
	if err != nil {
		return false, fmt.Sprintf("cannot tell whether it is an art track (%v)", err)
	}
	if ArtTrackFromPage(page) {
		logf("[official] %s: an auto-generated track (its description says so)", videoID)
		return true, ""
	}
	return false, "not an auto-generated track (its description is not \"Provided to YouTube by\")"
}
