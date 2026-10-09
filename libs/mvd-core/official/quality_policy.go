package official

// Quality is the best a video offers to download: the tallest picture, in pixels, and
// the best audio, in kilobits a second. Each is zero when there is none.
type Quality struct {
	Height    int
	AudioKbps int
}

const (
	// pictureCeiling and audioCeiling are where more no longer counts for more: a 4K
	// picture and a 320 kbps track are as good as anyone needs.
	pictureCeiling = 2160
	audioCeiling   = 320

	// qualityMargin is how much better (on a scale of 0 to 1) an upload must be to be
	// taken instead of the playlist's own: not worth a different file for less.
	qualityMargin = 0.1
)

// Score rates a quality from 0 to 1, the picture counting for more than the audio
// since a music video's picture is what varies most between uploads.
func (q Quality) Score() float64 {
	picture := float64(min(q.Height, pictureCeiling)) / pictureCeiling
	audio := float64(min(q.AudioKbps, audioCeiling)) / audioCeiling
	return 0.7*picture + 0.3*audio
}
