package ytdlp

import (
	"encoding/json"
	"fmt"
	"math"
)

// MediaQuality is the best a video offers to download: the tallest picture, in
// pixels, and the best audio, in kilobits a second. Each is zero when there is none.
type MediaQuality struct {
	Height    int
	AudioKbps int
}

// ParseMediaQuality reads the formats of `yt-dlp -J` output for a video.
func ParseMediaQuality(data []byte) (MediaQuality, error) {
	var info struct {
		Formats []struct {
			Height *float64 `json:"height"`
			VCodec string   `json:"vcodec"`
			ACodec string   `json:"acodec"`
			ABR    *float64 `json:"abr"`
			TBR    *float64 `json:"tbr"`
		} `json:"formats"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return MediaQuality{}, fmt.Errorf("cannot parse the formats of the video: %w", err)
	}

	var quality MediaQuality
	for _, f := range info.Formats {
		hasVideo := f.VCodec != "" && f.VCodec != "none"
		hasAudio := f.ACodec != "" && f.ACodec != "none"
		if hasVideo && f.Height != nil {
			quality.Height = max(quality.Height, int(*f.Height))
		}
		if !hasAudio {
			continue
		}
		// A muxed format's total rate includes its picture, so it only stands for the
		// audio when the format is audio alone.
		switch {
		case f.ABR != nil:
			quality.AudioKbps = max(quality.AudioKbps, int(math.Round(*f.ABR)))
		case !hasVideo && f.TBR != nil:
			quality.AudioKbps = max(quality.AudioKbps, int(math.Round(*f.TBR)))
		}
	}
	return quality, nil
}
