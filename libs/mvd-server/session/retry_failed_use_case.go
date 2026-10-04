package session

import "youtube-downloader/libs/mvd-core/engine"

// RetryFailed queues everything that has failed so far again: every failed entry, and
// every link that could not be looked up (unless it has since worked). It says how many
// were queued. It is what to do after something that was behind the failures has been
// fixed, such as an out-of-date downloader being replaced.
func (s *Session) RetryFailed() int {
	eng := s.engine()
	if eng == nil {
		return 0
	}

	s.mu.Lock()
	var failed []int
	for _, entry := range s.state.Entries {
		// An id is reserved before its entry arrives, so there can be a gap.
		if entry != nil && entry.State == engine.StateFailed {
			failed = append(failed, entry.ID)
		}
	}
	var unlisted []string
	for _, playlist := range s.state.Playlists {
		if playlist.Err != "" {
			unlisted = append(unlisted, playlist.URL)
		}
	}
	s.mu.Unlock()

	retried := 0
	for _, id := range failed {
		if eng.Retry(id) {
			retried++
		}
	}
	// Adding a link again is how one that failed to be listed is tried again. Add leaves
	// alone any that has since been listed fine, and a closed session just adds nothing.
	if result, err := s.Add(unlisted); err == nil {
		retried += len(result.Added)
	}

	return retried
}
