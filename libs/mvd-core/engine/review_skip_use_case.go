package engine

// Skip rejects a song held for review (Options.HoldForReview), also one that was accepted and has not
// begun to download: it is not downloaded, and the run can end without it. It returns false when the
// song is not waiting for review.
func (e *Engine) Skip(entryID int) bool {
	e.mu.Lock()
	if !e.opts.HoldForReview || entryID < 0 || entryID >= len(e.entries) {
		e.mu.Unlock()
		return false
	}
	en := e.entries[entryID]
	if e.removed[en.info.Playlist] || (en.state != StatePlanned && en.state != StateNotFound && en.state != StateFailed && en.state != StateAccepted) {
		e.mu.Unlock()
		return false
	}
	e.mu.Unlock()
	e.setState(en, StateSkipped, "")
	e.log(en.info.Playlist, entryID, "cancelled by you")

	return true
}
