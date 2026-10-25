package downloadsession

import (
	"fmt"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/songlist"
	"youtube-downloader/libs/mvd-core/viewport"
)

// VisibleEntries are the ids of the songs shown for the playlist the cursor is on: the ones the
// filter lets through, in the order the person chose.
func (s Session) VisibleEntries() []int {
	if s.SelPlaylist >= len(s.State.Playlists) {
		return nil
	}
	ids := s.State.Playlists[s.SelPlaylist].Entries
	if s.Filter.Active() {
		ids = nil
		for _, id := range s.State.Playlists[s.SelPlaylist].Entries {
			if s.Filter.Matches(s.State.Entries[id]) {
				ids = append(ids, id)
			}
		}
	}

	return s.Order.Sorted(ids, s.State.Entries)
}

// SelectedEntry is the song the cursor is on, nil when there is none.
func (s Session) SelectedEntry() *runstate.Entry {
	ids := s.VisibleEntries()
	if len(ids) == 0 {
		return nil
	}

	return s.State.Entries[ids[viewport.Clamp(s.SelEntry, 0, len(ids)-1)]]
}

// ActiveEntry is the id of the first song coming in now, -1 when none is.
func (s Session) ActiveEntry() int {
	for _, en := range s.State.Entries {
		if en != nil && (en.State == engine.StateDownloading || en.State == engine.StateMerging || en.State == engine.StateResolving) {
			return en.ID
		}
	}

	return -1
}

// OutputLines are the log lines of what the cursor is on: the song, else the playlist.
func (s Session) OutputLines() []string {
	if en := s.SelectedEntry(); s.Pane != PaneLists && en != nil {
		return en.Log
	}
	if s.SelPlaylist < len(s.State.Playlists) {
		return s.State.Playlists[s.SelPlaylist].Log
	}

	return nil
}

// JumpTo puts the cursor on song id.
func (s *Session) JumpTo(id int) {
	en := s.State.Entry(id)
	if en == nil {
		return
	}
	s.SelPlaylist = max(s.State.PlaylistPos(en.Playlist), 0)
	for i, eid := range s.VisibleEntries() {
		if eid == id {
			s.SelEntry = i

			break
		}
	}
	s.OutScroll = 0
}

// Move goes delta rows in the pane the person is in: through the playlists, the songs, or back
// and forth in the output.
func (s *Session) Move(delta int) {
	switch s.Pane {
	case PaneLists:
		s.Follow = false
		s.SelPlaylist = viewport.Clamp(s.SelPlaylist+delta, 0, len(s.State.Playlists)-1)
		s.SelEntry = 0
		s.OutScroll = 0
	case PaneEntries:
		s.Follow = false
		s.SelEntry = viewport.Clamp(s.SelEntry+delta, 0, len(s.VisibleEntries())-1)
		s.OutScroll = 0
	case PaneOutput:
		s.OutScroll = viewport.Clamp(s.OutScroll-delta, 0, max(0, len(s.OutputLines())-1))
	}
}

// ScrollOutput scrolls the output by delta lines, wherever the person is.
func (s *Session) ScrollOutput(delta int) {
	pane := s.Pane
	s.Pane = PaneOutput
	s.Move(delta)
	s.Pane = pane
}

// NextPane goes to the next pane; outputShown says whether the renderer has room for the output.
func (s *Session) NextPane(outputShown bool) {
	s.Pane = (s.Pane + 1) % paneCount
	if s.Pane == PaneOutput && !outputShown && !s.FullLog {
		s.Pane = PaneLists
	}
}

// PreviousPane goes to the previous pane.
func (s *Session) PreviousPane(outputShown bool) {
	s.Pane = (s.Pane + paneCount - 1) % paneCount
	if s.Pane == PaneOutput && !outputShown && !s.FullLog {
		s.Pane = PaneEntries
	}
}

// FocusPane puts the person in a pane.
func (s *Session) FocusPane(p Pane) { s.Pane = p }

// ChoosePlaylist puts the cursor on playlist i, as a click does: the person is in the lists and
// the cursor stops following the downloads.
func (s *Session) ChoosePlaylist(i int) {
	s.Pane = PaneLists
	s.Follow = false
	s.SelPlaylist, s.SelEntry, s.OutScroll = i, 0, 0
}

// ChooseEntry puts the cursor on the song at row i of the songs shown, as a click does. A playlist
// that could not be read or is not listed yet has no songs to put it on.
func (s *Session) ChooseEntry(i int) {
	s.Pane = PaneEntries
	if s.SelPlaylist >= len(s.State.Playlists) {
		return
	}
	if pl := s.State.Playlists[s.SelPlaylist]; pl.Err != "" || !pl.Listed {
		return
	}
	s.Follow = false
	s.SelEntry, s.OutScroll = i, 0
}

// SelectPlaylist puts the cursor on playlist i, as far as there is one, leaving the rest as it is.
func (s *Session) SelectPlaylist(i int) {
	s.SelPlaylist = max(min(i, len(s.State.Playlists)-1), 0)
}

// ToggleFollow keeps the cursor on the song coming in, or lets it go.
func (s *Session) ToggleFollow() {
	s.Follow = !s.Follow
	if !s.Follow {
		s.Say("follow off")

		return
	}
	s.OutScroll = 0
	if id := s.ActiveEntry(); id >= 0 {
		s.JumpTo(id)
	}
	s.Say("following active download")
}

// ToggleTechnical shows the log of the tool in the details, or the plain account again.
func (s *Session) ToggleTechnical() {
	s.Technical = !s.Technical
	s.OutScroll = 0
	if s.Technical {
		s.Say("technical log on: t goes back to the plain details")
	} else {
		s.Say("plain details on")
	}
}

// ToggleFullLog gives the log the whole screen, or takes it back.
func (s *Session) ToggleFullLog(outputShown bool) {
	s.FullLog = !s.FullLog
	switch {
	case s.FullLog:
		s.Technical = true // the full screen is for reading the log
		s.Pane = PaneOutput
	case s.Pane == PaneOutput && !outputShown:
		s.Pane = PaneEntries
	}
}

// ToggleFailedOnly shows the failed songs only, or all of them again.
func (s *Session) ToggleFailedOnly() {
	s.Filter = s.Filter.Only(songlist.GroupFailed)
	s.SelEntry = 0
	if s.Filter.Active() {
		s.Say("showing failed songs only")
	} else {
		s.Say("showing all songs")
	}
}

// ToggleStatus shows or hides the songs of a group of statuses.
func (s *Session) ToggleStatus(group string) {
	s.Filter = s.Filter.Toggled(group)
	s.SelEntry = 0
}

// ClearFilter shows every song.
func (s *Session) ClearFilter() {
	s.Filter = songlist.Filter{}
}

// SortBy sorts the songs by a column; the column that sorts already turns its direction round. The
// song the cursor is on stays the one it is on.
func (s *Session) SortBy(column string) {
	keep := -1
	if en := s.SelectedEntry(); en != nil {
		keep = en.ID
	}
	s.Order = s.Order.Clicked(column)
	for i, id := range s.VisibleEntries() {
		if id == keep {
			s.SelEntry = i
		}
	}
}

// Retry tries a failed song again, or the failed songs of the playlist when that is what the
// person is in.
func (s *Session) Retry() {
	if s.Ctrl == nil {
		s.Say("nothing is running; start the downloads first")

		return
	}
	if s.Pane == PaneEntries {
		en := s.SelectedEntry()
		if en == nil {
			return
		}
		if en.State != engine.StateFailed {
			s.Say("only failed entries can be retried")

			return
		}
		if s.Ctrl.Retry(en.ID) {
			s.Say(fmt.Sprintf("retrying %02d %s", en.Index, songlist.VideoTitle(en)))
		}

		return
	}
	if s.SelPlaylist < len(s.State.Playlists) {
		n := s.Ctrl.RetryPlaylist(s.State.Playlists[s.SelPlaylist].Index)
		s.Say(fmt.Sprintf("re-queued %d failed entries", n))
	}
}
