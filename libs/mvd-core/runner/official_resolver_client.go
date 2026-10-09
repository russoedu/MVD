package runner

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/official"
)

// engineResolver adapts the official video resolver to the port the engine
// uses, which tells an official video from an upload that is only better.
type engineResolver struct {
	*official.Resolver
}

func (r engineResolver) ResolveVersion(videoID, title, channel string, durationSec int, logf func(format string, a ...interface{})) engine.Resolution {
	version := r.Resolver.ResolveVersion(videoID, title, channel, durationSec, logf)
	return engine.Resolution{VideoID: version.ID, Official: version.Official, Reason: version.Reason}
}
