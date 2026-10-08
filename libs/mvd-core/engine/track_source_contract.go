package engine

import "context"

// Track is a song of a playlist from another service. It has no YouTube video
// yet: the engine finds one when it gets to the track.
type Track struct {
	Title string
	// Artist is the artists as the service lists them, comma separated.
	Artist     string
	DurationMs int
}

// TrackSource is the port through which the engine takes playlists that
// YouTube does not host (Spotify, for now). The engine never imports the
// slices that implement it.
type TrackSource interface {
	// Handles reports whether the source can list this link.
	Handles(link string) bool
	// Tracks lists the playlist behind the link.
	Tracks(ctx context.Context, link string) (title string, tracks []Track, err error)
	// Find returns the YouTube video for a track, and whether it is the official
	// one. An empty id means nothing matched. It logs through logf.
	Find(ctx context.Context, track Track, logf func(format string, a ...interface{})) (videoID string, official bool, err error)
}
