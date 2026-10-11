package outline

import "youtube-downloader/libs/mvd-core/keybar"

// KeysOf is the outline of a key bar: the same hints the bar draws.
func KeysOf(hints []keybar.Hint) []Key {
	keys := make([]Key, len(hints))
	for i, h := range hints {
		keys[i] = Key{Key: h.Key, Description: h.Desc}
	}
	return keys
}
