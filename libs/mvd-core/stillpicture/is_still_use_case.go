package stillpicture

import (
	"image"
	"sync"
)

// Checker says whether a video is a still picture, remembering what it found since
// the same video is asked about by several songs of a playlist.
type Checker struct {
	Frames interface {
		Frames(videoID string) ([]image.Image, error)
	}

	mu    sync.Mutex
	known map[string]bool
}

// NewChecker returns a checker that reads the frames from YouTube.
func NewChecker() *Checker {
	return &Checker{Frames: NewFramesClient()}
}

// IsStill reports whether the video is the same picture throughout. An error means the
// frames could not be read, which says nothing about the video.
func (c *Checker) IsStill(videoID string) (bool, error) {
	c.mu.Lock()
	still, ok := c.known[videoID]
	c.mu.Unlock()
	if ok {
		return still, nil
	}

	frames, err := c.Frames.Frames(videoID)
	if err != nil {
		return false, err
	}
	still = looksStill(frames)

	c.mu.Lock()
	if c.known == nil {
		c.known = map[string]bool{}
	}
	c.known[videoID] = still
	c.mu.Unlock()
	return still, nil
}
