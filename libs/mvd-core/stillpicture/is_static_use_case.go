package stillpicture

import (
	"net/http"
	"sync"
)

// minStoryboardFrames is how many frames a storyboard must give to be trusted over the
// three frames of the video.
const minStoryboardFrames = 20

// Checker says whether a video is static (a picture with a song over it, not a video
// that moves), remembering what it found since the same video is asked about by several
// songs of a playlist.
type Checker struct {
	HTTP *http.Client
	// Storyboard returns the storyboard of a video (the app reads it from yt-dlp's
	// description of the video). Optional: without it, or when it fails, the three
	// frames YouTube makes of every video are compared instead.
	Storyboard func(videoID string) (Storyboard, error)
	// FrameURL is the address of frame n (1 to 3) of a video.
	FrameURL func(videoID string, n int) string

	mu    sync.Mutex
	known map[string]bool
}

// NewChecker returns a checker that reads from YouTube. storyboard may be nil.
func NewChecker(storyboard func(videoID string) (Storyboard, error)) *Checker {
	return &Checker{HTTP: newHTTPClient(), Storyboard: storyboard, FrameURL: framesURL}
}

// IsStatic reports whether the video does not really move. An error means nothing could
// be read, which says nothing about the video.
func (c *Checker) IsStatic(videoID string) (bool, error) {
	c.mu.Lock()
	static, ok := c.known[videoID]
	c.mu.Unlock()
	if ok {
		return static, nil
	}

	frames, err := c.frames(videoID)
	if err != nil {
		return false, err
	}
	static = !hasMotion(frames)

	c.mu.Lock()
	if c.known == nil {
		c.known = map[string]bool{}
	}
	c.known[videoID] = static
	c.mu.Unlock()
	return static, nil
}

// frames reads the storyboard, or the three frames when there is none.
func (c *Checker) frames(videoID string) ([]greyFrame, error) {
	if c.Storyboard != nil {
		if sb, err := c.Storyboard(videoID); err == nil {
			if frames, err := readStoryboard(c.HTTP, sb); err == nil && len(frames) >= minStoryboardFrames {
				return frames, nil
			}
		}
	}
	return readThreeFrames(c.HTTP, c.FrameURL, videoID)
}
