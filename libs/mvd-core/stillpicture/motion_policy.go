package stillpicture

import "sort"

const (
	// minFrames is how many frames are needed to say anything.
	minFrames = 3
	// movingMedian is the median share of the frame that changes between one frame and
	// the next, from which a video moves. Measured on a sample of 30 uploads: real
	// videos 0.12 to 0.43, art tracks and covers 0.00 to 0.02, lyric videos over one
	// picture 0.00 to 0.06.
	movingMedian = 0.10
	// farGap is how many frames apart two are compared as well, so that a slow
	// movement, which changes little between neighbours, still shows.
	farGap = 5
	// farChange is the share of the frame that must have changed between two frames
	// farGap apart for the pair to count as a change.
	farChange = 0.10
	// movingFarShare is the share of those pairs that must show a change, from which
	// a video moves although its median is lower: a performance with long stills.
	movingFarShare = 0.5
)

// hasMotion says whether a video, given as frames spread over its length, really
// moves: either most of its frames differ a good deal from the next, or most of them
// differ a good deal from the one farGap places on.
func hasMotion(frames []greyFrame) bool {
	if len(frames) < minFrames {
		return false
	}
	structures := make([]greyFrame, len(frames))
	for i, frame := range frames {
		structures[i] = structureOf(frame)
	}

	next := make([]float64, 0, len(frames)-1)
	for i := 0; i+1 < len(structures); i++ {
		next = append(next, changedShare(structures[i], structures[i+1]))
	}
	if median(next) >= movingMedian {
		return true
	}

	far := 0
	pairs := 0
	for i := 0; i+farGap < len(structures); i++ {
		pairs++
		if changedShare(structures[i], structures[i+farGap]) > farChange {
			far++
		}
	}
	return pairs > 0 && float64(far)/float64(pairs) >= movingFarShare
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	return sorted[len(sorted)/2]
}
