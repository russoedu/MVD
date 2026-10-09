package main

// minKnownSongs is the fewest songs of the baseline a check needs to find in the playlist
// to say anything: with fewer, a few removed videos would decide the result.
const minKnownSongs = 20

// Verdict is whether YouTube still gives the answers that were recorded.
type Verdict struct {
	// Known is how many songs of the playlist have a recorded answer.
	Known int
	// Different are the known songs whose answer is not the recorded one.
	Different []Outcome
	// New are songs of the playlist that have no recorded answer yet.
	New []Outcome
	// Gone is how many recorded songs are no longer in the playlist.
	Gone int
	// Failed says too many answers changed, or there were too few to tell.
	Failed bool
	// Reason says why it failed, empty when it did not.
	Reason string
}

// Judge compares the answers now with the recorded ones. It fails when more than
// maxDifferent (a share, 0 to 1) of the known songs differ: one video removed or blocked
// is not a change of YouTube, a tenth of the playlist answering differently is.
func Judge(baseline Baseline, outcomes []Outcome, maxDifferent float64) Verdict {
	verdict := Verdict{}
	inPlaylist := map[string]bool{}
	for _, o := range outcomes {
		inPlaylist[o.ID] = true
		if !o.Known {
			verdict.New = append(verdict.New, o)
			continue
		}
		verdict.Known++
		if o.Found != o.Expected {
			verdict.Different = append(verdict.Different, o)
		}
	}
	for _, song := range baseline.Songs {
		if !inPlaylist[song.ID] {
			verdict.Gone++
		}
	}

	switch {
	case verdict.Known < minKnownSongs:
		verdict.Failed = true
		verdict.Reason = "too few songs of the baseline were found in the playlist to tell whether YouTube changed"
	case float64(len(verdict.Different)) > maxDifferent*float64(verdict.Known):
		verdict.Failed = true
		verdict.Reason = "more of the songs than allowed now give a different answer"
	}

	return verdict
}
