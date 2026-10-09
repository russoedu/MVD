package stillpicture

import (
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"net/http"
	"time"
)

// Storyboard describes the sheets of small frames YouTube keeps of a video.
type Storyboard struct {
	// Width and Height are the size of one frame, Rows and Columns how many are on a sheet.
	Width, Height, Rows, Columns int
	// Fragments are the sheets in order, each with the seconds of video it covers.
	Fragments []StoryboardFragment
}

// StoryboardFragment is one sheet of a storyboard.
type StoryboardFragment struct {
	URL         string
	DurationSec float64
}

// maxSheets bounds how many sheets are read, however long the video is.
const maxSheets = 6

// readStoryboard fetches the sheets of a storyboard and cuts them into frames. The last
// sheet is padded with empty tiles past the end of the video; the number of real frames
// is worked out from the seconds the sheets cover.
func readStoryboard(client *http.Client, sb Storyboard) ([]greyFrame, error) {
	perSheet := sb.Rows * sb.Columns
	if perSheet == 0 || len(sb.Fragments) == 0 || sb.Width == 0 || sb.Height == 0 {
		return nil, fmt.Errorf("the storyboard describes no frames")
	}

	total := 0.0
	for _, fragment := range sb.Fragments {
		total += fragment.DurationSec
	}
	interval := sb.Fragments[0].DurationSec / float64(perSheet)
	real := len(sb.Fragments) * perSheet
	if interval > 0 && total > 0 {
		real = min(real, int(math.Round(total/interval)))
	}

	var frames []greyFrame
	for i, fragment := range sb.Fragments {
		if i == maxSheets {
			break
		}
		sheet, err := fetchSheet(client, fragment.URL)
		if err != nil {
			return nil, err
		}
		frames = append(frames, cutSheet(sheet, sb.Width, sb.Height, sb.Rows, sb.Columns, min(perSheet, real-len(frames)))...)
	}
	return frames, nil
}

func fetchSheet(client *http.Client, url string) (image.Image, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storyboard sheet: the image server answered %s", resp.Status)
	}
	return jpeg.Decode(resp.Body)
}

// newHTTPClient is the client the real checker reads storyboards and frames with.
func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Second}
}
