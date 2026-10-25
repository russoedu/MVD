package review

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
)

// ApplyEvent takes an event of the plan run. over is true when the run has gone idle, which is the
// last event the host need wait for.
func (s *Session) ApplyEvent(ev interface{}) (effects []Effect, over bool) {
	if !s.Planning {
		return nil, true
	}
	s.PlanState.Apply(ev)
	if _, idle := ev.(engine.EvIdle); idle {
		return s.planFinished(), true
	}

	return nil, false
}

// planFinished takes the plan from the run that made it, once every song is dealt with.
func (s *Session) planFinished() []Effect {
	// A question about last time's review is waiting: the plan is taken when it is answered.
	if s.Mode == AskPrevious {
		s.planDone = true

		return nil
	}
	if s.planner != nil {
		fresh := playlisteditor.FromPlan(s.planner.Plan())
		if len(s.previous) > 0 {
			fresh = playlisteditor.Merge(fresh, s.previous)
		}
		s.Rows = fresh
	}
	s.Planning, s.planDone = false, false
	s.previous, s.Saved = nil, nil
	s.Selected = 0
	s.autosave()
	if len(s.Rows) == 0 {
		s.Flash("no songs were found in the list")
	} else {
		s.Flash(fmt.Sprintf("%d songs planned: check them, then download", len(s.Rows)))
	}

	return []Effect{PlanFinished{}}
}

// AnswerPrevious answers the question about the review of last time: bring it along, or leave it.
func (s *Session) AnswerPrevious(bring bool) []Effect {
	if bring {
		s.previous = s.Saved
	} else {
		s.previous = nil
	}
	s.Mode, s.Saved = Browsing, nil
	if s.planDone {
		return s.planFinished()
	}

	return nil
}

// MarkDownloaded marks the songs with these keys as downloaded, so the next download skips them.
func (s *Session) MarkDownloaded(keys []string) {
	n := playlisteditor.MarkDownloaded(s.Rows, keys...)
	s.autosave()
	switch n {
	case 0:
		s.Flash("nothing was downloaded")
	case 1:
		s.Flash("1 song downloaded; it will be skipped next time")
	default:
		s.Flash(fmt.Sprintf("%d songs downloaded; they will be skipped next time", n))
	}
}

// PendingCount is how many songs of the saved review of last time are not downloaded.
func (s Session) PendingCount() int {
	pending := 0
	for _, r := range s.Saved {
		if !r.Downloaded {
			pending++
		}
	}

	return pending
}

// PlanProgress is how many songs the plan run has dealt with, out of how many it knows of.
func (s Session) PlanProgress() (done, total int) {
	t := s.PlanState.Tally()

	return t.Planned + t.NotFound + t.Failed, t.Total
}
