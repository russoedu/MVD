package engine

// Release sends a song held for review (Options.HoldForReview) to the download stage.
// videoID is the video to download in place of the proposal, "" to take the proposal as
// it is. It returns false when the song is not waiting for review or has nothing to
// download. The song is fixed from then on: a retry downloads the same video and does
// not look for another.
func (e *Engine) Release(entryID int, videoID string) bool {
	e.mu.Lock()
	if !e.opts.HoldForReview || entryID < 0 || entryID >= len(e.entries) {
		e.mu.Unlock()
		return false
	}
	en := e.entries[entryID]
	if e.removed[en.info.Playlist] || (en.state != StatePlanned && en.state != StateSkipped) || (en.targetID == "" && videoID == "") {
		e.mu.Unlock()
		return false
	}
	if videoID != "" && videoID != en.targetID {
		en.targetID = videoID
		en.official, en.better, en.ownOfficial = false, false, false
		en.userChose = true
		en.reason = "chosen by you"
	}
	en.fixed = true
	en.state = StateAccepted
	en.err = ""
	e.pending++
	e.idleSent = false
	e.mu.Unlock()

	e.emit(EvEntryState{Entry: entryID, State: StateAccepted, TargetID: en.targetID, Official: en.official, Better: en.better, OwnOfficial: en.ownOfficial, Reason: en.reason})
	e.log(en.info.Playlist, entryID, "released for download")
	e.queue.Push(entryID)
	return true
}

// ReleaseAll releases every song held for review as it was proposed, and returns how
// many were released.
func (e *Engine) ReleaseAll() int {
	e.mu.Lock()
	var ids []int
	for _, en := range e.entries {
		if en.state == StatePlanned && !e.removed[en.info.Playlist] {
			ids = append(ids, en.info.ID)
		}
	}
	e.mu.Unlock()
	n := 0
	for _, id := range ids {
		if e.Release(id, "") {
			n++
		}
	}
	return n
}
