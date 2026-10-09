package engine

// OfficialResolver is the port through which the engine swaps an
// auto-generated art track for its official music video. The official
// slice implements it; the engine never imports that slice.
type OfficialResolver interface {
	// Wanted reports whether an upload is worth a lookup at all (cheap, no
	// network). It must not rely on the channel name: auto-generated art tracks
	// often show the artist's own name.
	Wanted(title, channel, uploader string) bool
	// ResolveVersion looks for a better version of the upload to download: the
	// official video, or failing that the upload of the song with the best
	// quality, logging progress through logf. The title, channel and length (in
	// seconds, 0 when unknown) of the upload let it search for the song by name.
	ResolveVersion(videoID, title, channel string, durationSec int, logf func(format string, a ...interface{})) Resolution
}

// Resolution is the answer of an OfficialResolver.
type Resolution struct {
	// VideoID is the version to download instead, "" when the upload stays.
	VideoID string
	// Official is true when VideoID is the official video, false when it is only the
	// upload of the song with the best quality.
	Official bool
	// Reason is a short explanation for the log.
	Reason string
}
