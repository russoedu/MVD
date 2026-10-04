// Package api is the HTTP surface a browser uses to watch and feed the run.
//
// It is JSON over HTTP, plus one server-sent-events stream, and it is meant to
// be mounted under /api on a server that listens on the loopback address only.
package api

import (
	"context"

	"youtube-downloader/libs/mvd-server/session"
	"youtube-downloader/libs/mvd-server/snapshot"
)

// Sessions is what the handlers need from the run. *session.Session satisfies it;
// the interface is here so the handlers can be tested without an engine.
type Sessions interface {
	// Snapshot is the run as it is now.
	Snapshot() snapshot.Snapshot
	// WaitSince blocks until the snapshot is newer than version, or ctx ends.
	WaitSince(ctx context.Context, version int64) (snapshot.Snapshot, bool)
	// Add queues URLs and says what became of each.
	Add(raw []string) (session.AddResult, error)
	// Retry re-queues one failed entry.
	Retry(entry int) bool
	// RetryPlaylist re-queues a playlist's failed entries and says how many.
	RetryPlaylist(playlist int) int
}
