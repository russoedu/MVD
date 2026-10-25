package engine

// Pause stops new downloads from starting. The ones under way finish, and everything
// else waits (queued) until Resume. Lookups of songs go on, so what is waiting is ready.
func (e *Engine) Pause() { e.throttle.setPaused(true) }

// Resume lets downloads start again after Pause.
func (e *Engine) Resume() { e.throttle.setPaused(false) }

// Paused reports whether Pause is in force.
func (e *Engine) Paused() bool {
	e.throttle.mu.Lock()
	defer e.throttle.mu.Unlock()
	return e.throttle.paused
}

// RetryAll queues every failed song again, as the person asked (the retry cycles are not
// counted against it), and returns how many it queued.
func (e *Engine) RetryAll() int {
	e.mu.Lock()
	var ids []int
	for _, en := range e.entries {
		if en.state == StateFailed && !e.removed[en.info.Playlist] {
			ids = append(ids, en.info.ID)
		}
	}
	e.mu.Unlock()
	n := 0
	for _, id := range ids {
		if e.Retry(id) {
			n++
		}
	}
	return n
}
