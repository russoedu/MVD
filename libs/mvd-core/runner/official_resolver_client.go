package runner

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/official"
)

// engineResolver adapts the official video resolver to the port the engine uses: a
// lookup in two stages (naming, then picking) whose answer tells an official video
// from an upload that is only better.
type engineResolver struct {
	*official.Resolver
}

func (r engineResolver) Identify(videoID, title, channel string, durationSec int, logf func(format string, a ...interface{})) engine.Identification {
	identified := r.Resolver.Identify(videoID, title, channel, durationSec, logf)
	artist, name, _ := identified.Named()
	return engine.Identification{Done: identified.Done, Resolution: resolutionOf(identified.Version), Artist: artist, Title: name, Album: identified.Album(), State: identified}
}

func (r engineResolver) Pick(identification engine.Identification, logf func(format string, a ...interface{})) engine.Resolution {
	identified, ok := identification.State.(official.Identified)
	if !ok {
		return engine.Resolution{}
	}
	return resolutionOf(r.Resolver.Pick(identified, logf))
}

// PickWithout searches again for the song, leaving out the videos that were found wanting.
func (r engineResolver) PickWithout(identification engine.Identification, rejected []string, logf func(format string, a ...interface{})) engine.Resolution {
	identified, ok := identification.State.(official.Identified)
	if !ok {
		return engine.Resolution{}
	}
	return resolutionOf(r.Resolver.PickWithout(identified, rejected, logf))
}

func resolutionOf(version official.Version) engine.Resolution {
	return engine.Resolution{VideoID: version.ID, Official: version.Official, OwnOfficial: version.OwnOfficial, Reason: version.Reason}
}
