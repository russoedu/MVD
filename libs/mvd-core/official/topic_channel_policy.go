// Package official finds the official music video that YouTube links from
// the "Music" card in the description of an auto-generated art track.
package official

import "strings"

// IsTopicChannel reports whether a channel name is one of YouTube's
// auto-generated "<Artist> - Topic" channels (the art-track uploads).
func IsTopicChannel(name string) bool {
	return strings.HasSuffix(strings.TrimSpace(name), " - Topic")
}
