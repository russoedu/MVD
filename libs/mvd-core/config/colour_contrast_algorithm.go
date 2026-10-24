package config

import (
	"math"
	"strconv"
	"strings"
)

// luminance is the relative luminance of a #rgb or #rrggbb colour (WCAG 2.x), and false
// when it is not a colour.
func luminance(colour string) (float64, bool) {
	c := strings.TrimPrefix(strings.TrimSpace(colour), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) != 6 {
		return 0, false
	}
	n, err := strconv.ParseUint(c, 16, 32)
	if err != nil {
		return 0, false
	}
	channel := func(v uint64) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(n>>16&0xff) + 0.7152*channel(n>>8&0xff) + 0.0722*channel(n&0xff), true
}

// Contrast is the WCAG contrast ratio between two colours, from 1 (the same) to 21; 0 when
// either is not a colour. Text of normal size needs 4.5.
func Contrast(a, b string) float64 {
	la, okA := luminance(a)
	lb, okB := luminance(b)
	if !okA || !okB {
		return 0
	}
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

// MinTextContrast is the contrast small text needs against its background.
const MinTextContrast = 4.5
