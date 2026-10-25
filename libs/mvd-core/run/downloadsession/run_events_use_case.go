package downloadsession

import (
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

// Apply takes in what the run reported: the state is updated, a link that was just added becomes
// the one the person looks at, and with Follow on the cursor goes to the song that begins.
func (s *Session) Apply(ev interface{}) {
	id := s.State.Apply(ev)
	if added, ok := ev.(engine.EvPlaylistAdded); ok {
		// A link that was just added is the one the person looks at: its songs come in under it.
		if at := s.State.PlaylistPos(added.Source.Index); at >= 0 {
			s.SelPlaylist, s.SelEntry = at, 0
		}
	}
	s.SelPlaylist = max(min(s.SelPlaylist, len(s.State.Playlists)-1), 0)
	if !s.Loaded && (s.State.Idle || s.State.AllListed()) {
		s.Loaded = true
	}
	if st, ok := ev.(engine.EvEntryState); ok && s.Follow {
		if st.State == engine.StateDownloading || st.State == engine.StateResolving {
			s.JumpTo(id)
		}
	}
	if lg, ok := ev.(engine.EvLog); ok && s.OutScroll > 0 {
		// Keep the view anchored while the user is reading older lines.
		if en := s.SelectedEntry(); en != nil && lg.Entry == en.ID {
			s.OutScroll++
		}
	}
}

// RunClosed takes in that the run's events are over. A hosted screen keeps what the run reported
// and lets the run go; any other leaves.
func (s *Session) RunClosed() Effect {
	if s.Hosted {
		s.Ctrl = nil

		return RunEnded{}
	}

	return Quit{}
}

// Show shows state, for a list that is not being run: what an earlier session reported.
func (s *Session) Show(state *runstate.State) {
	s.State = state
	s.SelPlaylist, s.SelEntry, s.OutScroll = 0, 0, 0
}

// Attach shows a run that has begun: the screen starts again from what the run reports.
func (s *Session) Attach(c Controller) {
	s.Ctrl = c
	s.State = runstate.New(c.Sources())
	s.Loaded = len(s.State.Playlists) == 0
	s.Follow = true
	// The link the cursor was on stays the one, when the run has it: a link was just added there.
	s.SelPlaylist = max(min(s.SelPlaylist, len(s.State.Playlists)-1), 0)
	s.SelEntry, s.OutScroll = 0, 0
}

// Detach lets the run go, keeping what it reported.
func (s *Session) Detach() {
	s.Ctrl, s.Paused = nil, false
}
