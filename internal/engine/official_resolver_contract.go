package engine

// OfficialResolver is the port through which the engine swaps an
// auto-generated art track for its official music video. The official
// slice implements it; the engine never imports that slice.
type OfficialResolver interface {
	// Wanted reports whether an upload from this channel should be resolved.
	Wanted(channel, uploader string) bool
	// ResolveLog returns the official video id ("" when none was found) and
	// a short reason, logging progress through logf.
	ResolveLog(videoID string, logf func(format string, a ...interface{})) (string, string)
}
