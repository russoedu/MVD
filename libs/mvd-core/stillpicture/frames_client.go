package stillpicture

import (
	"fmt"
	"image/jpeg"
	"net/http"
)

// framesURL is the address of frame n (1 to 3) of a video: the frames YouTube makes
// of every video at 25, 50 and 75 percent of its length, which exist even when there
// is no storyboard.
func framesURL(videoID string, n int) string {
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/%d.jpg", videoID, n)
}

// readThreeFrames fetches the three frames of a video. It fails when any cannot be read.
func readThreeFrames(client *http.Client, url func(videoID string, n int) string, videoID string) ([]greyFrame, error) {
	frames := make([]greyFrame, 0, 3)
	for n := 1; n <= 3; n++ {
		frame, err := readFrame(client, url(videoID, n), videoID, n)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

func readFrame(client *http.Client, url, videoID string, n int) (greyFrame, error) {
	resp, err := client.Get(url)
	if err != nil {
		return greyFrame{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return greyFrame{}, fmt.Errorf("frame %d of %s: the image server answered %s", n, videoID, resp.Status)
	}
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return greyFrame{}, fmt.Errorf("frame %d of %s: %w", n, videoID, err)
	}
	return greyOf(img), nil
}
