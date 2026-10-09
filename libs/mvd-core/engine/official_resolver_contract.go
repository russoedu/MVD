package engine

// OfficialResolver is the port through which the engine swaps an
// auto-generated art track for its official music video. The official
// slice implements it; the engine never imports that slice.
type OfficialResolver interface {
	// Wanted reports whether an upload is worth a lookup at all (cheap, no
	// network). It must not rely on the channel name: auto-generated art tracks
	// often show the artist's own name.
	Wanted(title, channel, uploader string) bool
	// Identify is the first half of the lookup for a better version of an upload:
	// it tells art tracks from videos, follows the link of the description to the
	// official video and names the song in the music databases. It may already know
	// the answer.
	Identify(videoID, title, channel string, durationSec int, logf func(format string, a ...interface{})) Identification
	// Pick is the second half: it searches for the official video of the named song
	// and, failing that, the upload of it with the best quality.
	Pick(identification Identification, logf func(format string, a ...interface{})) Resolution
}

// Identification is what Identify learned about an upload, handed to Pick.
type Identification struct {
	// Done is true when Resolution is already the answer.
	Done       bool
	Resolution Resolution
	// Artist and Title are the song as a music database named it, when it did.
	Artist, Title string
	// State is the resolver's own notes for Pick; the engine only carries it.
	State any
}

// Resolution is the answer of an OfficialResolver.
type Resolution struct {
	// VideoID is the version to download instead, "" when the upload stays.
	VideoID string
	// Official is true when VideoID is the official video, false when it is only the
	// upload of the song with the best quality.
	Official bool
	// OwnOfficial is true when the upload itself is the official video, so VideoID is
	// empty and nothing better is needed.
	OwnOfficial bool
	// Reason is a short explanation for the log.
	Reason string
}
