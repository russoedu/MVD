package official

import (
	"fmt"
	"sort"
)

// maxQualityChecked is how many uploads of a song have their formats looked up: each
// costs a yt-dlp run.
const maxQualityChecked = 3

// fromBestQuality is what is left when no upload says it is the official video: the
// upload of the song with the best picture and sound, among the most viewed ones the
// searches saw and the playlist's own. It returns the id of the best, or "" when the
// playlist's own is as good as any (or the quality cannot be looked up).
func (r *Resolver) fromBestQuality(song Song, results []SearchResult, logf func(format string, a ...interface{})) (string, string) {
	if r.Sources.Quality == nil {
		return "", ""
	}

	var uploads []SearchResult
	seen := map[string]bool{}
	for _, res := range results {
		if !seen[res.ID] && isUploadOfTheSong(res, song) {
			seen[res.ID] = true
			uploads = append(uploads, res)
		}
	}
	if len(uploads) == 0 {
		return "", ""
	}
	sort.SliceStable(uploads, func(i, j int) bool { return uploads[i].Views > uploads[j].Views })

	// An art track is no better than the playlist's own, and YouTube shows many under the
	// artist's own name, so YouTube Music's tag is asked of each before it is taken.
	var chosen []SearchResult
	for _, res := range uploads {
		if len(chosen) == maxQualityChecked {
			break
		}
		kind := res.MusicType
		if kind == "" && r.Sources.Type != nil {
			kind, _ = r.Sources.Type(res.ID)
		}
		if artTrack, _ := ArtTrackFromVideoType(kind); artTrack {
			continue
		}
		// Nor is a picture with the song over it a better version of the song.
		if r.isStill(res.ID, song.OwnID, logf) {
			continue
		}
		chosen = append(chosen, res)
	}
	uploads = chosen

	own := Quality{}
	if song.OwnID != "" {
		if quality, err := r.Sources.Quality(song.OwnID); err == nil {
			own = quality
			// A still picture has no picture worth counting, whatever size the file is.
			if song.OwnIsStill || r.isStill(song.OwnID, song.OwnID, logf) {
				own.Height = 0
			}
		}
	}

	best, bestQuality, bestScore := SearchResult{}, Quality{}, -1.0
	for _, res := range uploads {
		quality, err := r.Sources.Quality(res.ID)
		if err != nil {
			logf("[official] %s: cannot look up the quality of %s: %v", song.OwnID, res.ID, err)
			continue
		}
		if score := quality.Score(); score > bestScore {
			best, bestQuality, bestScore = res, quality, score
		}
	}
	if bestScore < 0 || bestScore <= own.Score()+qualityMargin {
		return "", ""
	}
	return best.ID, fmt.Sprintf("no official video found; the best quality is %s by %q (%dp, %d kbps, against %dp, %d kbps)",
		best.ID, best.Channel, bestQuality.Height, bestQuality.AudioKbps, own.Height, own.AudioKbps)
}

// isStill says whether a video is a picture with the song over it. When that cannot be
// told it is not, so a failing image server never costs a good version.
func (r *Resolver) isStill(videoID, ownID string, logf func(format string, a ...interface{})) bool {
	if r.Sources.Still == nil || videoID == "" {
		return false
	}
	still, err := r.Sources.Still(videoID)
	if err != nil {
		logf("[official] %s: cannot tell whether %s is a still picture: %v", ownID, videoID, err)
		return false
	}
	if still {
		logf("[official] %s: %s is only a still picture, left out", ownID, videoID)
	}
	return still
}
