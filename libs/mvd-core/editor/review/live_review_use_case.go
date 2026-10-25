package review

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Refresh takes the songs of the live run as they stand, and puts the decisions on them.
func (s *Session) Refresh() {
	snap := s.live.Snapshot()
	rows := make([]playlistfile.Entry, len(snap))
	states := make([]engine.EntryState, len(snap))
	for i, sn := range snap {
		row := playlisteditor.FromPlan([]engine.PlannedEntry{sn.PlannedEntry})[0]
		if d, ok := s.decisions[row.Key()]; ok {
			row.Decision, row.ChosenID, row.ChosenTitle = d.Decision, d.ChosenID, d.ChosenTitle
		}
		row.Downloaded = sn.State == engine.StateDone || sn.State == engine.StateDuplicate
		rows[i], states[i] = row, sn.State
	}
	s.Rows, s.States = rows, states
	if s.Selected >= len(rows) {
		s.Selected = max(len(rows)-1, 0)
	}
}

// releaseFor sends the selected song of a live review to download when the decision is one that
// downloads it, and refuses a decision the run cannot follow: a song still being looked up, or one
// that has already been sent.
func (s *Session) releaseFor(decision, chosenID string) bool {
	state := engine.StatePlanned
	if s.Selected < len(s.States) {
		state = s.States[s.Selected]
	}
	held := state == engine.StatePlanned || state == engine.StateNotFound || state == engine.StateFailed
	switch decision {
	case playlistfile.DecisionNone, playlistfile.DecisionSkipped:
		if !held {
			s.Flash("this song has been sent to download already")

			return false
		}

		return true
	}
	if state == engine.StateQueued || state == engine.StateResolving {
		s.Flash("this song is still being looked up")

		return false
	}
	if !held {
		s.Flash("this song has been sent to download already")

		return false
	}
	video := chosenID
	if decision == playlistfile.DecisionOriginal {
		video = s.Rows[s.Selected].Upload.ID
	}
	if !s.live.Release(s.Selected, video) {
		s.Flash("this song cannot be downloaded: replace it or skip it")

		return false
	}
	s.States[s.Selected] = engine.StateQueued

	return true
}

// remember keeps the decision on row i, so the next refresh puts it back.
func (s *Session) remember(i int) {
	if s.decisions == nil {
		s.decisions = map[string]playlistfile.Entry{}
	}
	s.decisions[s.Rows[i].Key()] = s.Rows[i]
	s.Version++
}
