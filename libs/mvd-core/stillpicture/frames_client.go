package stillpicture

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"time"
)

// FramesClient fetches the three frames YouTube keeps of a video.
type FramesClient struct {
	HTTP *http.Client
	// URL is the address of frame n (1 to 3) of a video.
	URL func(videoID string, n int) string
}

// NewFramesClient returns a client pointed at the real YouTube image server.
func NewFramesClient() *FramesClient {
	return &FramesClient{
		HTTP: &http.Client{Timeout: 15 * time.Second},
		URL: func(videoID string, n int) string {
			return fmt.Sprintf("https://i.ytimg.com/vi/%s/%d.jpg", videoID, n)
		},
	}
}

// Frames returns the frames of a video at 25, 50 and 75 percent of its length. It
// fails when any of them cannot be read.
func (c *FramesClient) Frames(videoID string) ([]image.Image, error) {
	frames := make([]image.Image, 0, 3)
	for n := 1; n <= 3; n++ {
		frame, err := c.frame(videoID, n)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

func (c *FramesClient) frame(videoID string, n int) (image.Image, error) {
	resp, err := c.HTTP.Get(c.URL(videoID, n))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frame %d of %s: the image server answered %s", n, videoID, resp.Status)
	}
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("frame %d of %s: %w", n, videoID, err)
	}
	return img, nil
}
