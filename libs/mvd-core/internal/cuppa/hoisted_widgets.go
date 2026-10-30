// Written by the Go export of Cuppa and maintained by hand since: change it here.

package cuppa

import (
	"strconv"
	"strings"
)

// Float returns the number in the prop key, read as a decimal.
// It returns fallback when the prop is missing or is not a number.
func (p Props) Float(key string, fallback float64) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(p[key]), 64)
	if err != nil {
		return fallback
	}
	return f
}
