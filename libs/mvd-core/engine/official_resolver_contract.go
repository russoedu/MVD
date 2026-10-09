package engine

// OfficialResolver is the port through which the engine swaps an
// auto-generated art track for its official music video. The official
// slice implements it; the engine never imports that slice.
type OfficialResolver interface {
	// Wanted reports whether an upload is worth a lookup at all (cheap, no
	// network). It must not rely on the channel name: auto-generated art tracks
	// often show the artist's own name.
	Wanted(title, channel, uploader string) bool
	// ResolveLog returns the official video id ("" when none was found) and
	// a short reason, logging progress through logf. The title and channel
	// of the art track let it search for the video by name.
	ResolveLog(videoID, title, channel string, logf func(format string, a ...interface{})) (string, string)
}
