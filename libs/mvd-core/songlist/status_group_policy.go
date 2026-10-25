package songlist

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// The groups of statuses the songs can be filtered by: what the person sees of a song, not every
// state of the engine.
const (
	GroupOfficial    = "official"
	GroupOriginal    = "original"
	GroupAccepted    = "accepted"
	GroupWaiting     = "waiting"
	GroupDownloading = "downloading"
	GroupSaved       = "saved"
	GroupHad         = "had"
	GroupMissing     = "missing"
	GroupRejected    = "rejected"
	GroupFailed      = "failed"
)

// Group is one group of statuses and what it is called.
type Group struct{ Key, Label string }

// Groups are the groups of statuses in the order a filter lists them.
var Groups = []Group{
	{GroupOfficial, "Official found"}, {GroupOriginal, "Original only"}, {GroupAccepted, "Accepted"}, {GroupWaiting, "Waiting"}, {GroupDownloading, "Downloading"}, {GroupSaved, "Saved"},
	{GroupHad, "Already had"}, {GroupMissing, "No official video"}, {GroupRejected, "Rejected"}, {GroupFailed, "Failed"},
}

// GroupOf is the group of statuses a song is in.
func GroupOf(en *runstate.Entry) string {
	switch en.State {
	case engine.StatePlanned:
		if FoundOfficial(en) {
			return GroupOfficial
		}

		return GroupOriginal
	case engine.StateAccepted:
		return GroupAccepted
	case engine.StateQueued:
		return GroupWaiting
	case engine.StateResolving, engine.StateDownloading, engine.StateMerging:
		return GroupDownloading
	case engine.StateDone:
		return GroupSaved
	case engine.StateDuplicate:
		return GroupHad
	case engine.StateNotFound:
		return GroupMissing
	case engine.StateSkipped:
		return GroupRejected
	case engine.StateFailed:
		return GroupFailed
	}

	return ""
}

// FoundOfficial says whether what was found for a song that waits is an official video: the one
// found for it, or its own upload when that is the official one.
func FoundOfficial(en *runstate.Entry) bool { return en.Official || en.OwnOfficial }
