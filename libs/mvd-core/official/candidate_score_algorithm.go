package official

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Song is what is known about the song to find a video for.
type Song struct {
	Title string
	// Artists are the names to accept in a video's title or channel; the first is
	// the one searched for.
	Artists []string
	// DurationSec is the length of the song's audio, or 0 when unknown. The
	// video of a song is often a different cut, so it only nudges the score.
	DurationSec int
	// OwnID is the upload the song came from, never a candidate.
	OwnID string
}

// Kind says what a candidate is.
type Kind int

const (
	// KindVideo is a music video.
	KindVideo Kind = iota
	// KindAudio is an official audio or lyric upload: the same song as the art
	// track but from the artist's own channel.
	KindAudio
)

func (k Kind) String() string {
	if k == KindAudio {
		return "official audio"
	}
	return "official video"
}

// Pick is the candidate chosen for a song.
type Pick struct {
	ID      string
	Kind    Kind
	Score   float64
	Channel string
	// Why lists what counted for the candidate, for the log.
	Why string
}

// minTitleSimilarity is how much of the song's title a video's title must
// carry for the video to be the song at all.
const minTitleSimilarity = 0.6

// minFanViews is how many views an upload from a channel that is not the
// artist's needs to be trusted: a tiny re-upload that says "official" is more
// often a rip of something else than the video itself.
const minFanViews = 50_000

// cutsToReject are cuts that are never the song's video unless the song itself
// is that cut. Lyric and audio uploads are not here: they are a kind of their own.
var cutsToReject = []string{
	"cover", "karaoke", "live", "reaction", "instrumental", "remix", "slowed",
	"sped up", "8d", "nightcore", "reverb", "tutorial",
}

var audioMarkers = []string{"audio", "lyric", "lyrics", "visualizer", "visualiser"}

func hasAny(title, original string, words []string) bool {
	for _, word := range words {
		if containsWords(title, word) && !containsWords(original, word) {
			return true
		}
	}
	return false
}

// scoreCandidate scores one search result for a song. ok is false when the
// result cannot be the song's video at all.
func scoreCandidate(res SearchResult, song Song) (Pick, bool) {
	if res.ID == "" || res.ID == song.OwnID || IsTopicChannel(res.Channel) {
		return Pick{}, false
	}

	resTitle := normalize(res.Title)
	original := normalize(song.Title)
	if hasAny(resTitle, original, cutsToReject) {
		return Pick{}, false
	}

	similarity := TitleSimilarity(song.Title, res.Title, song.Artists)
	if similarity < minTitleSimilarity {
		return Pick{}, false
	}

	artistInTitle, artistChannel := len(song.Artists) == 0, false
	for _, artist := range song.Artists {
		if band := normalize(artist); band != "" && containsWords(resTitle, band) {
			artistInTitle = true
		}
		// A verified channel that carries the artist's name is theirs too ("Deep Purple
		// Official", "Michael Sembello (The Master)").
		if ChannelIsArtist(res.Channel, artist) || (res.Verified && containsWords(normalize(res.Channel), normalize(artist))) {
			artistChannel = true
		}
	}
	if !artistInTitle && !artistChannel {
		return Pick{}, false
	}
	if !artistChannel && !res.Verified && res.Views > 0 && res.Views < minFanViews {
		return Pick{}, false
	}

	kind := KindVideo
	if hasAny(resTitle, original, audioMarkers) {
		kind = KindAudio
	}
	official := containsWords(resTitle, "official")

	// A video is taken when it says it is official or when the artist's own
	// channel uploaded it (Nickelback's "How You Remind Me" never says it). An
	// audio or lyric upload is taken only from the artist's own channel.
	if kind == KindVideo && !official && !artistChannel {
		return Pick{}, false
	}
	if kind == KindAudio && !artistChannel {
		return Pick{}, false
	}

	var why []string
	score := similarity * 4
	why = append(why, fmt.Sprintf("title %.0f%%", similarity*100))
	if artistChannel {
		score += 3
		why = append(why, "artist's channel")
	} else {
		why = append(why, "artist in title")
	}
	if official {
		score += 2
		why = append(why, "says official")
	}
	if res.Verified {
		score++
		why = append(why, "verified")
	}
	if res.Views > 0 {
		score += math.Min(math.Log10(float64(res.Views))/8, 1)
	}
	if note, bonus := durationFit(song.DurationSec, res.Duration); note != "" {
		score += bonus
		why = append(why, note)
	}
	// A video beats an audio upload however well the audio scores.
	if kind == KindAudio {
		score -= 20
	}

	return Pick{ID: res.ID, Kind: kind, Score: score, Channel: res.Channel, Why: strings.Join(why, ", ")}, true
}

// durationFit compares the lengths of the song and of a video. A music video
// often adds an intro, so a different length only costs a little.
func durationFit(song, video int) (string, float64) {
	if song <= 0 || video <= 0 {
		return "", 0
	}
	difference := math.Abs(float64(video-song)) / float64(song)
	switch {
	case difference <= 0.05:
		return "same length", 1.5
	case difference <= 0.15:
		return fmt.Sprintf("length %+d%%", int(math.Round((float64(video)/float64(song)-1)*100))), 0.8
	case difference <= 0.35:
		return fmt.Sprintf("length %+d%%", int(math.Round((float64(video)/float64(song)-1)*100))), 0.2
	default:
		return fmt.Sprintf("length %+d%%", int(math.Round((float64(video)/float64(song)-1)*100))), -0.8
	}
}

// RankCandidates scores every result that can be the song's video, best first.
func RankCandidates(results []SearchResult, song Song) []Pick {
	var picks []Pick
	seen := map[string]bool{}
	for _, res := range results {
		if seen[res.ID] {
			continue
		}
		seen[res.ID] = true
		if pick, ok := scoreCandidate(res, song); ok {
			picks = append(picks, pick)
		}
	}
	sort.SliceStable(picks, func(i, j int) bool { return picks[i].Score > picks[j].Score })
	return picks
}

// PickBestVideo chooses the video of a song among search results, or reports
// none. Music videos win over official audio, the artist's own channel and an
// official title over the rest, a similar length over a different one.
func PickBestVideo(results []SearchResult, song Song) (Pick, bool) {
	picks := RankCandidates(results, song)
	if len(picks) == 0 {
		return Pick{}, false
	}
	return picks[0], true
}
