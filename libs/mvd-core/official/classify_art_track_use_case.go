package official

import "fmt"

// isArtTrack decides whether an upload is an auto-generated art track, the only
// kind worth looking an official video up for, and returns what YouTube Music
// knows of it. Many art tracks show the artist's own name as the channel
// instead of "<Artist> - Topic", so the name says nothing: YouTube Music's tag
// is asked first, then the channel name, then the description on the watch
// page. The returned reason says why an upload was left alone.
//
// Without a TrackDescriber every upload is treated as an art track, as before.
func (r *Resolver) isArtTrack(videoID, channel string, logf func(format string, a ...interface{})) (bool, TrackInfo, string) {
	if r.TrackInfos == nil {
		return true, TrackInfo{}, ""
	}

	info, err := r.TrackInfos(videoID)
	if err != nil {
		logf("[official] %s: YouTube Music could not say what this upload is (%v)", videoID, err)
	}
	if artTrack, known := ArtTrackFromVideoType(info.Type); known {
		if artTrack {
			logf("[official] %s: an auto-generated track (YouTube Music: %s, %q)", videoID, info.Type, info.Artist)
			return true, info, ""
		}
		return false, info, fmt.Sprintf("already a video, not an art track (YouTube Music: %s)", info.Type)
	}

	if IsTopicChannel(channel) {
		return true, info, ""
	}
	page, err := r.get(r.WatchBase + videoID)
	if err != nil {
		return false, info, fmt.Sprintf("cannot tell whether it is an art track (%v)", err)
	}
	if ArtTrackFromPage(page) {
		logf("[official] %s: an auto-generated track (its description says so)", videoID)
		return true, info, ""
	}
	return false, info, "not an auto-generated track (its description is not \"Provided to YouTube by\")"
}
