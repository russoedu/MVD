package downloadsession

import "youtube-downloader/libs/mvd-core/engine"

// Controller is what the download screen needs from the engine.
type Controller interface {
	Events() <-chan interface{}
	Sources() []engine.PlaylistSource
	LogPath() string
	Retry(entryID int) bool
	RetryPlaylist(playlist int) int
}

// Chooser is what a run offers when it holds each song for the person to choose: a song waits,
// found and proposed, until it is accepted, with the video it proposes or another, or rejected. A
// song that was rejected can be accepted again. The engine does when it is told to hold the songs.
type Chooser interface {
	// Release accepts song entry: it goes to download as proposed when videoID is "", else with that
	// video.
	Release(entry int, videoID string) bool
	// ReleaseAll accepts every song that waits, as proposed, and says how many.
	ReleaseAll() int
	// Skip rejects song entry: it is not downloaded.
	Skip(entry int) bool
	// Swap gives song entry another video to download, also when it was saved already: what it
	// saved is then taken away. False when the song is on its way and cannot be swapped.
	Swap(entry int, videoID string) bool
}

// SourceAdder is what a run offers when it can take more links while it works.
type SourceAdder interface {
	// AddSource queues another playlist or video; false once the run has ended.
	AddSource(url string) (engine.PlaylistSource, bool)
}

// SourceRemover is what a run offers when it takes a link out while it goes on.
type SourceRemover interface {
	// RemoveSource takes playlist (the engine's number for it) out of the run.
	RemoveSource(playlist int) bool
}
