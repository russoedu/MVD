package selfupdate

import (
	"strconv"
	"strings"
)

// parseVersion reads x.y.z; ok is false for anything else, a development build included.
func parseVersion(s string) (v [3]int, ok bool) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, false
		}
		v[i] = n
	}

	return v, true
}

// isNewer is true when latest is a higher version than have. A build that has no number (a
// development build) is never offered an update.
func isNewer(have, latest string) bool {
	h, ok := parseVersion(have)
	if !ok {
		return false
	}
	l, ok := parseVersion(latest)
	if !ok {
		return false
	}
	for i := range h {
		if l[i] != h[i] {
			return l[i] > h[i]
		}
	}

	return false
}
