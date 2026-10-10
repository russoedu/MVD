package audiofix

import (
	"math"
	"strconv"
	"strings"
)

const (
	// silentBelow is the level, in dB, under which a channel is taken as silent: a
	// channel of dither or encoder noise sits around -90, a quiet song well above -50.
	silentBelow = -70.0
	// audibleAbove is the level the other channel must have, so a track that is quiet
	// as a whole is never taken for a one-sided one.
	audibleAbove = -45.0
)

// ParseChannelLevels reads the RMS level of each channel from ffmpeg's astats report,
// in channel order. A silent channel is reported as -inf.
func ParseChannelLevels(report string) []float64 {
	var levels []float64
	for _, line := range strings.Split(report, "\n") {
		_, after, found := strings.Cut(line, "RMS level dB:")
		if !found {
			continue
		}
		text := strings.TrimSpace(after)
		switch {
		case text == "-inf":
			levels = append(levels, math.Inf(-1))
		default:
			if value, err := strconv.ParseFloat(text, 64); err == nil {
				levels = append(levels, value)
			}
		}
	}
	return levels
}

// SilentChannel says which channel of a stereo track is silent while the other one
// carries the sound: 0 for the left, 1 for the right. ok is false for a track that
// plays on both sides, a silent track, or one that is not stereo.
func SilentChannel(levels []float64) (channel int, ok bool) {
	if len(levels) != 2 {
		return 0, false
	}
	left, right := levels[0], levels[1]
	switch {
	case left < silentBelow && right > audibleAbove:
		return 0, true
	case right < silentBelow && left > audibleAbove:
		return 1, true
	}
	return 0, false
}
