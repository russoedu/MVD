// Package session owns the one long-lived run that a front end watches and feeds.
//
// A terminal run is a batch: build an engine for a list, watch it finish. A tray
// app does not end, so something has to hold the engine, mirror its events into
// a state a browser can read, and take new URLs while it works. That is this.
package session

import (
	"context"

	"youtube-downloader/libs/mvd-core/engine"
)

// Engine is the part of *engine.Engine a session drives. It is an interface so
// the session can be tested without yt-dlp: the real engine satisfies it as is.
type Engine interface {
	// Run lists and downloads until ctx is cancelled, then closes Events.
	Run(ctx context.Context)
	// Events carries everything a renderer needs to know.
	Events() <-chan interface{}
	// Sources are the playlists known so far.
	Sources() []engine.PlaylistSource
	// AddSource queues another playlist; false once Run has finished.
	AddSource(url string) (engine.PlaylistSource, bool)
	// Retry re-queues a failed entry.
	Retry(entryID int) bool
	// RetryPlaylist re-queues every failed entry of a playlist.
	RetryPlaylist(playlist int) int
}

// Factory builds an engine for the first URLs, ready to Run. It is called once,
// when the first URLs arrive, and not at start-up: building the engine can fetch
// browser cookies, which wants a URL to check them against.
type Factory func(ctx context.Context, urls []string) (Engine, error)
