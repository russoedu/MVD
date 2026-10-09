package ytdlp

import (
	"encoding/json"
	"fmt"
)

// Storyboard is the set of sheets of small frames YouTube keeps of a video.
type Storyboard struct {
	Width, Height, Rows, Columns int
	Fragments                    []StoryboardFragment
}

// StoryboardFragment is one sheet, with the seconds of video it covers.
type StoryboardFragment struct {
	URL         string
	DurationSec float64
}

// VideoProbe is what `yt-dlp -J` says of a video that is useful for choosing between
// uploads: the best it offers to download, and its storyboard.
type VideoProbe struct {
	Quality    MediaQuality
	Storyboard Storyboard
}

// ParseVideoProbe reads the output of `yt-dlp -J` for a video.
func ParseVideoProbe(data []byte) (VideoProbe, error) {
	quality, err := ParseMediaQuality(data)
	if err != nil {
		return VideoProbe{}, err
	}

	var info struct {
		Formats []struct {
			FormatID  string `json:"format_id"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			Rows      int    `json:"rows"`
			Columns   int    `json:"columns"`
			Fragments []struct {
				URL      string  `json:"url"`
				Duration float64 `json:"duration"`
			} `json:"fragments"`
		} `json:"formats"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return VideoProbe{}, fmt.Errorf("cannot parse the storyboard of the video: %w", err)
	}

	// The storyboard with a hundred frames to a sheet and the biggest frames: a couple
	// of small sheets cover a whole video.
	probe := VideoProbe{Quality: quality}
	for _, f := range info.Formats {
		if len(f.FormatID) < 2 || f.FormatID[:2] != "sb" || f.Rows*f.Columns != 100 || len(f.Fragments) == 0 {
			continue
		}
		if f.Width <= probe.Storyboard.Width {
			continue
		}
		board := Storyboard{Width: f.Width, Height: f.Height, Rows: f.Rows, Columns: f.Columns}
		for _, fragment := range f.Fragments {
			board.Fragments = append(board.Fragments, StoryboardFragment{URL: fragment.URL, DurationSec: fragment.Duration})
		}
		probe.Storyboard = board
	}
	return probe, nil
}
