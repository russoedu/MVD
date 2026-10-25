package downloadsnapshot

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/run/downloadsession"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
	"youtube-downloader/libs/mvd-core/wording"
)

// HeaderOf is the counters of the header and the label on its top edge.
func HeaderOf(s downloadsession.Session) Header {
	t := s.State.Tally()
	official, original := plannedCounts(s)
	moving := t.Running > 0 && !s.State.Idle && !s.Paused

	// What matters is always there, in everyday words; the rest only when it has something to say.
	downloadingTone := songlist.ToneDim
	if moving {
		downloadingTone = songlist.ToneHighlight
	}
	counters := []Counter{
		{Kind: CounterWaiting, Label: "Waiting", Count: t.Queued, Tone: songlist.ToneText},
		{Kind: CounterDownloading, Label: "Downloading", Count: t.Running, Tone: downloadingTone, Moving: moving},
		{Kind: CounterDone, Label: "Done", Count: t.Done, Tone: songlist.ToneSuccess},
		{Kind: CounterFailed, Label: "Failed", Count: t.Failed + unreadableLinks(s), Tone: songlist.ToneError},
	}
	for _, extra := range []Counter{
		{Kind: CounterOfficialFound, Label: "Official found", Count: official, Tone: songlist.ToneSuccess},
		{Kind: CounterOriginalOnly, Label: "Original only", Count: original, Tone: songlist.ToneHighlight},
		{Kind: CounterAccepted, Label: "Accepted", Count: t.Accepted, Tone: songlist.ToneFocus},
		{Kind: CounterOfficialVideo, Label: "Official video", Count: t.Official, Tone: songlist.ToneFocus},
		{Kind: CounterBetterQuality, Label: "Better quality", Count: t.Better, Tone: songlist.ToneFocus},
		{Kind: CounterAlreadyHad, Label: "Already had", Count: t.Duplicate, Tone: songlist.ToneDim},
		{Kind: CounterNoOfficial, Label: "No official video", Count: t.NotFound, Tone: songlist.ToneDim},
		{Kind: CounterRejected, Label: "Rejected", Count: t.Skipped, Tone: songlist.ToneDim},
		{Kind: CounterTriedAgain, Label: "Tried again", Count: t.Retried, Tone: songlist.ToneFocus},
	} {
		if extra.Count > 0 {
			counters = append(counters, extra)
		}
	}

	return Header{Counters: counters, Right: stateLabel(s, t)}
}

// stateLabel is where the run is, and how long it has taken.
func stateLabel(s downloadsession.Session, t runstate.Tally) string {
	elapsed := runstate.HumanDuration(s.State.Elapsed())
	switch {
	case s.Editable:
		return "" // nothing is running: there is no time to count
	case s.Paused && t.Accepted > 0:
		return "ACCEPTED · press Download to start · " + elapsed
	case s.State.Idle && t.Planned > 0:
		return "CHOOSE WHAT TO DOWNLOAD · " + elapsed
	case s.Paused:
		return "STOPPED · " + elapsed
	case s.State.Idle:
		return "FINISHED · " + elapsed
	case t.Downloading > 0:
		return downloadingSummary(t) + " · " + elapsed
	}

	return elapsed
}

// downloadingSummary says how many songs are coming in now and how fast, all of them together:
// "Downloading 4 songs · 12.4MiB/s". The speed is left out while it is not known.
func downloadingSummary(t runstate.Tally) string {
	summary := "Downloading " + wording.Counted(t.Downloading, "song")
	if t.Speed > 0 {
		summary += " · " + runstate.HumanSpeed(t.Speed)
	}

	return summary
}

// plannedCounts are how many songs wait for the person: with an official video found for them, and
// with only the original.
func plannedCounts(s downloadsession.Session) (official, original int) {
	for _, en := range s.State.Entries {
		if en == nil || en.State != engine.StatePlanned {
			continue
		}
		if songlist.FoundOfficial(en) {
			official++
		} else {
			original++
		}
	}

	return official, original
}

// unreadableLinks is how many links could not be read at all: no song of them is counted, and for
// the person it is a failure all the same.
func unreadableLinks(s downloadsession.Session) int {
	n := 0
	for _, pl := range s.State.Playlists {
		if pl.Err != "" {
			n++
		}
	}

	return n
}
