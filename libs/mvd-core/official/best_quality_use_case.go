package official

import (
	"fmt"
	"slices"
	"sort"
	"youtube-downloader/libs/mvd-core/official/trackpick"
)

// maxQualityChecked is how many uploads of a song have their formats looked up: each
// costs a yt-dlp run.
const maxQualityChecked = 5

// fromBestQuality is what is left when no upload says it is the official video: the
// upload of the song with the best picture and sound, among the most viewed ones the
// searches saw and the playlist's own. It returns the id of the best, or "" when the
// playlist's own is as good as any (or the quality cannot be looked up).
func (r *Resolver) fromBestQuality(song trackpick.Song, results []trackpick.SearchResult, logf func(format string, a ...interface{})) (string, string) {
	if r.Sources.Quality == nil {
		return "", ""
	}

	var uploads []trackpick.SearchResult
	seen := map[string]bool{}
	for _, res := range results {
		if !seen[res.ID] && trackpick.IsUploadOfTheSong(res, song) && !slices.Contains(song.Rejected, res.ID) {
			seen[res.ID] = true
			uploads = append(uploads, res)
		}
	}
	if len(uploads) == 0 {
		return "", ""
	}
	// The most viewed come first, but an upload that says in its title that it is in high
	// quality is looked at before them: a real HD video often has few views.
	sort.SliceStable(uploads, func(i, j int) bool {
		hi, hj := trackpick.AdvertisesQuality(uploads[i].Title), trackpick.AdvertisesQuality(uploads[j].Title)
		if hi != hj {
			return hi
		}
		return uploads[i].Views > uploads[j].Views
	})

	// An art track is no better than the playlist's own, and YouTube shows many under the
	// artist's own name, so YouTube Music's tag is asked of each before it is taken.
	var chosen []trackpick.SearchResult
	for _, res := range uploads {
		if len(chosen) == maxQualityChecked {
			break
		}
		kind := res.MusicType
		if kind == "" && r.Sources.Type != nil {
			kind, _ = r.Sources.Type(res.ID)
		}
		if artTrack, _ := trackpick.ArtTrackFromVideoType(kind); artTrack {
			continue
		}
		// Nor is a picture with the song over it, or a lyric video over one picture, a
		// better version of the song.
		if r.isStatic(res.ID, song.OwnID, logf) {
			continue
		}
		chosen = append(chosen, res)
	}
	uploads = chosen

	own := trackpick.Quality{}
	if song.OwnID != "" {
		if quality, err := r.Sources.Quality(song.OwnID); err == nil {
			own = quality
			// A video that does not move has no picture worth counting, whatever size the file is.
			if song.OwnIsStill || r.isStatic(song.OwnID, song.OwnID, logf) {
				own.Height = 0
			}
		}
	}

	best, bestQuality, bestScore := trackpick.SearchResult{}, trackpick.Quality{}, -1.0
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
	if bestScore < 0 || bestScore <= own.Score()+trackpick.QualityMargin {
		return "", ""
	}
	return best.ID, fmt.Sprintf("no official video found; the best quality is %s by %q (%dp, %d kbps, against %dp, %d kbps)",
		best.ID, best.Channel, bestQuality.Height, bestQuality.AudioKbps, own.Height, own.AudioKbps)
}

// isStatic says whether a video does not really move: a picture with the song over it. When that cannot be
// told it is not, so a failing image server never costs a good version.
func (r *Resolver) isStatic(videoID, ownID string, logf func(format string, a ...interface{})) bool {
	if r.Sources.Static == nil || videoID == "" {
		return false
	}
	still, err := r.Sources.Static(videoID)
	if err != nil {
		logf("[official] %s: cannot tell whether %s moves: %v", ownID, videoID, err)
		return false
	}
	if still {
		logf("[official] %s: %s does not really move, left out", ownID, videoID)
	}
	return still
}
