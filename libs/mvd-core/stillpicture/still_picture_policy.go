package stillpicture

import "image"

// maxStillDifference is the most two frames of a still picture may differ. Measured on
// real uploads: art tracks differ by 0 to 0.4 between frames, music videos by 30 to 90.
// A little room is left for the compression of a picture that fades in or has a clock.
const maxStillDifference = 3.0

// minFrames is how many frames are needed to say anything.
const minFrames = 2

// looksStill says whether a video's frames are the same picture: no two of them
// differ by more than maxStillDifference. With fewer than two frames it cannot say.
func looksStill(frames []image.Image) bool {
	if len(frames) < minFrames {
		return false
	}
	for i := 0; i < len(frames); i++ {
		for j := i + 1; j < len(frames); j++ {
			if frameDifference(frames[i], frames[j]) > maxStillDifference {
				return false
			}
		}
	}
	return true
}
