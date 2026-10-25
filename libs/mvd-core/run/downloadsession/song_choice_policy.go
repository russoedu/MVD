package downloadsession

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Waiting says whether a song waits for the person to choose it.
func Waiting(en *runstate.Entry) bool { return en != nil && en.State == engine.StatePlanned }

// Swappable says whether the video of a song can be swapped now: it waits, was saved, failed or was
// rejected, and is not on its way.
func Swappable(en *runstate.Entry) bool {
	if en == nil {
		return false
	}
	switch en.State {
	case engine.StatePlanned, engine.StateDone, engine.StateDuplicate, engine.StateFailed, engine.StateNotFound, engine.StateSkipped:
		return true
	}

	return false
}

// Rejectable says whether a song can be rejected now: it waits, or it was accepted and has not begun.
func Rejectable(en *runstate.Entry) bool {
	return en != nil && (en.State == engine.StatePlanned || en.State == engine.StateAccepted)
}

// Acceptable says whether a song can be accepted now: it waits, or it was rejected.
func Acceptable(en *runstate.Entry) bool {
	return en != nil && (en.State == engine.StatePlanned || (en.State == engine.StateSkipped && en.TargetID != ""))
}
