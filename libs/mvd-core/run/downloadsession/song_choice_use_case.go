package downloadsession

import (
	"fmt"
	"sort"
	"strings"

	"youtube-downloader/libs/mvd-core/playlisteditor"
	"youtube-downloader/libs/mvd-core/runstate"
	"youtube-downloader/libs/mvd-core/wording"
)

// chooser is the run as something that holds songs for choosing, when it is.
func (s Session) chooser() (Chooser, bool) {
	if s.Ctrl == nil {
		return nil, false
	}
	c, ok := s.Ctrl.(Chooser)

	return c, ok
}

// CanChoose says whether the run holds its songs for the person to choose.
func (s Session) CanChoose() bool {
	_, ok := s.chooser()

	return ok
}

// AnyWaiting says whether some song waits for the person to choose it.
func (s Session) AnyWaiting() bool {
	return s.CanChoose() && s.State.Tally().Planned > 0
}

// WaitingIn is how many songs of a link wait for the person to choose them.
func (s Session) WaitingIn(pl *runstate.Playlist) int {
	n := 0
	for _, id := range pl.Entries {
		if Waiting(s.State.Entries[id]) {
			n++
		}
	}

	return n
}

// TargetIDs are the songs the actions work on: the ones ticked, else the one the cursor is on.
func (s Session) TargetIDs() []int {
	if len(s.Picked) > 0 {
		var ids []int
		for id := range s.Picked {
			if s.State.Entry(id) != nil {
				ids = append(ids, id)
			}
		}
		sort.Ints(ids)

		return ids
	}
	if en := s.SelectedEntry(); en != nil {
		return []int{en.ID}
	}

	return nil
}

// PickState is how ticked the songs of the table are: 0 none, 1 some, 2 all.
func (s Session) PickState() int {
	ids := s.VisibleEntries()
	n := 0
	for _, id := range ids {
		if s.Picked[id] {
			n++
		}
	}
	switch {
	case n == 0:
		return 0
	case n == len(ids):
		return 2
	}

	return 1
}

// accepted is what the screen says after songs were accepted: they wait for the Download button, or
// they go on at once.
func (s Session) accepted(n int) string {
	if s.Paused {
		return wording.Counted(n, "song") + " accepted. Press Download to start them"
	}

	return wording.Counted(n, "song") + " accepted: they will download"
}

// pickOnly is a click on a song, away from its box: it alone is ticked, or when it was the only one
// ticked the click takes the tick away.
func (s *Session) pickOnly(id int) {
	if len(s.Picked) == 1 && s.Picked[id] {
		s.Picked = nil

		return
	}
	s.Picked = map[int]bool{id: true}
}

// togglePick ticks or unticks a song.
func (s *Session) togglePick(id int) {
	picked := make(map[int]bool, len(s.Picked)+1)
	for k := range s.Picked {
		picked[k] = true
	}
	if picked[id] {
		delete(picked, id)
	} else {
		picked[id] = true
	}
	s.Picked = picked
}

// TickAll ticks every song of the table, or unticks them all when they all are.
func (s *Session) TickAll() {
	ids := s.VisibleEntries()
	all := len(ids) > 0
	for _, id := range ids {
		if !s.Picked[id] {
			all = false
		}
	}
	picked := make(map[int]bool, len(s.Picked))
	for k := range s.Picked {
		picked[k] = true
	}
	for _, id := range ids {
		if all {
			delete(picked, id)
		} else {
			picked[id] = true
		}
	}
	s.Picked = picked
}

// pickRange ticks, or unticks, the songs from one song to another, in the order they are shown.
func (s *Session) pickRange(from, to int, on bool) {
	ids := s.VisibleEntries()
	a, b := -1, -1
	for i, id := range ids {
		if id == from {
			a = i
		}
		if id == to {
			b = i
		}
	}
	if a < 0 || b < 0 {
		return
	}
	if a > b {
		a, b = b, a
	}
	picked := make(map[int]bool, len(s.Picked)+b-a+1)
	for k := range s.Picked {
		picked[k] = true
	}
	for _, id := range ids[a : b+1] {
		if on {
			picked[id] = true
		} else {
			delete(picked, id)
		}
	}
	s.Picked = picked
}

// TickSong ticks or unticks a song, and with extend all the songs from the one ticked last, as a
// list of mail does.
func (s *Session) TickSong(id int, extend bool) {
	on := !s.Picked[id]
	if extend && s.AnchorSet {
		s.pickRange(s.Anchor, id, on)
	} else {
		s.togglePick(id)
	}
	s.Anchor, s.AnchorSet = id, true
}

// ChooseSong chooses a song away from its box: it alone is ticked, and with extend every song from
// the one ticked last to it.
func (s *Session) ChooseSong(id int, extend bool) {
	if extend && s.AnchorSet {
		s.pickRange(s.Anchor, id, true)

		return
	}
	s.pickOnly(id)
	s.Anchor, s.AnchorSet = id, true
}

// TickCursorSong ticks the song the cursor is on, when the person is in the songs.
func (s *Session) TickCursorSong() {
	if !s.CanChoose() {
		return
	}
	if en := s.SelectedEntry(); en != nil && s.Pane == PaneEntries {
		s.TickSong(en.ID, false)
	}
}

// ExtendPick moves the cursor by delta songs with Shift held, and ticks every song from where the
// range began to the one the cursor is on.
func (s *Session) ExtendPick(delta int) {
	if !s.CanChoose() || s.Pane != PaneEntries {
		return
	}
	en := s.SelectedEntry()
	if en == nil {
		return
	}
	if !s.AnchorSet {
		s.Anchor, s.AnchorSet = en.ID, true
	}
	s.Move(delta)
	if now := s.SelectedEntry(); now != nil {
		s.Picked = nil
		s.pickRange(s.Anchor, now.ID, true)
	}
}

// AcceptAll accepts every song that waits.
func (s *Session) AcceptAll() {
	chooser, ok := s.chooser()
	if !ok {
		return
	}
	s.Say(s.accepted(chooser.ReleaseAll()))
	s.Picked = nil
}

// Accept accepts the songs: on a link with nothing ticked, all its waiting songs; otherwise the
// ones ticked, else the one the cursor is on.
func (s *Session) Accept() {
	chooser, ok := s.chooser()
	if !ok {
		return
	}
	if s.Pane == PaneLists && len(s.Picked) == 0 {
		sent := 0
		if s.SelPlaylist < len(s.State.Playlists) {
			for _, id := range s.State.Playlists[s.SelPlaylist].Entries {
				if Waiting(s.State.Entries[id]) && chooser.Release(id, "") {
					sent++
				}
			}
		}
		s.Say(s.accepted(sent))

		return
	}
	ids := s.TargetIDs()
	if len(ids) == 0 {
		s.Say("choose a song first")

		return
	}
	done := 0
	for _, id := range ids {
		if Acceptable(s.State.Entry(id)) && chooser.Release(id, "") {
			done++
		}
	}
	if done == 0 {
		s.Say("nothing to accept: the song is on its way, or has no video")
	} else {
		s.Say(s.accepted(done))
	}
	s.Picked = nil
}

// Reject rejects the songs ticked, else the one the cursor is on.
func (s *Session) Reject() {
	chooser, ok := s.chooser()
	if !ok {
		return
	}
	ids := s.TargetIDs()
	if len(ids) == 0 {
		s.Say("choose a song first")

		return
	}
	done := 0
	for _, id := range ids {
		if Rejectable(s.State.Entry(id)) && chooser.Skip(id) {
			done++
		}
	}
	if done == 0 {
		s.Say("nothing to reject: the song is on its way or done")
	} else {
		s.Say(fmt.Sprintf("%s rejected: they will not download. Accept brings them back", wording.Counted(done, "song")))
	}
	s.Picked = nil
}

// OpenVideo asks the host to show the video of the one song chosen.
func (s *Session) OpenVideo() Effect {
	if !s.CanChoose() {
		return nil
	}
	ids := s.TargetIDs()
	if len(ids) == 0 {
		s.Say("choose a song first")

		return nil
	}
	if len(ids) != 1 {
		s.Say("open one video at a time")

		return nil
	}
	en := s.State.Entry(ids[0])
	video := en.TargetID
	if video == "" {
		video = en.VideoID
	}

	return OpenURL{URL: "https://www.youtube.com/watch?v=" + video}
}

// BeginSwap starts asking for the address of another video for the song chosen. Swapping says
// whether it began.
func (s *Session) BeginSwap() {
	if !s.CanChoose() {
		return
	}
	ids := s.TargetIDs()
	if len(ids) == 0 {
		s.Say("choose a song first")

		return
	}
	en := s.State.Entry(ids[0])
	switch {
	case len(ids) != 1:
		s.Say("swap the video of one song at a time")

		return
	case !Swappable(en):
		s.Say("this song is on its way: swap it when it is saved")

		return
	}
	s.Swapping, s.SwapText, s.SwapEntry = true, "", en.ID
}

// SetSwapText says what has been typed as the address of the other video.
func (s *Session) SetSwapText(text string) { s.SwapText = text }

// CancelSwap gives up on swapping.
func (s *Session) CancelSwap() { s.Swapping = false }

// CommitSwap sends the song to download with the video whose address was typed.
func (s *Session) CommitSwap() {
	s.Swapping = false
	id, ok := playlisteditor.VideoID(strings.TrimSpace(s.SwapText))
	if !ok {
		s.Say("that is not the address of a YouTube video")

		return
	}
	chooser, held := s.chooser()
	switch {
	case !held:
	case chooser.Swap(s.SwapEntry, id):
		s.Say("accepted with the video you chose")
	default:
		s.Say("this song cannot be changed now")
	}
}
