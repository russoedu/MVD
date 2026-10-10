package audiofix

import (
	"math"
	"testing"
)

const report = `[Parsed_astats_0 @ 0000022d87f8e4c0] Channel: 1
[Parsed_astats_0 @ 0000022d87f8e4c0] RMS level dB: -inf
[Parsed_astats_0 @ 0000022d87f8e4c0] Channel: 2
[Parsed_astats_0 @ 0000022d87f8e4c0] RMS level dB: -13.773746
`

func TestLevelsAreReadInChannelOrderWithSilenceAsMinusInfinity(t *testing.T) {
	got := ParseChannelLevels(report)
	if len(got) != 2 || !math.IsInf(got[0], -1) || math.Abs(got[1]+13.773746) > 1e-9 {
		t.Errorf("got %v", got)
	}
}

func TestOnlyAStereoTrackWithOneSilentSideIsOneSided(t *testing.T) {
	for name, c := range map[string]struct {
		levels []float64
		want   int
		ok     bool
	}{
		"left silent":      {[]float64{math.Inf(-1), -14}, 0, true},
		"right silent":     {[]float64{-16, -91}, 1, true},
		"both play":        {[]float64{-14, -15}, 0, false},
		"both silent":      {[]float64{math.Inf(-1), math.Inf(-1)}, 0, false},
		"quiet as a whole": {[]float64{-80, -60}, 0, false},
		"mono":             {[]float64{-14}, 0, false},
		"surround":         {[]float64{-14, -14, math.Inf(-1), -14, -14, -14}, 0, false},
	} {
		got, ok := SilentChannel(c.levels)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("%s: got %d, %v", name, got, ok)
		}
	}
}
