package session

// AddResult says what happened to each line of a paste, so the UI can tell the
// user instead of leaving them to guess why nothing started.
type AddResult struct {
	// Added were queued.
	Added []string `json:"added"`
	// Duplicates are already in the run (and did not fail), so were left alone.
	Duplicates []string `json:"duplicates"`
	// Rejected were not an http(s) URL.
	Rejected []string `json:"rejected"`
}

// Add queues URLs. The first ones build the engine; later ones join it while it
// runs. A URL already in the run is left alone, unless that playlist failed to
// list, in which case adding it again is how it is tried again.
//
// The duplicate check reads the state, which the events pump updates a moment
// after the engine accepts a URL, so the same URL sent twice within that moment
// can slip through. The UI disables its button while a request is open.
func (s *Session) Add(raw []string) (AddResult, error) {
	accepted, rejected := NormalizeURLs(raw)
	result := AddResult{Added: []string{}, Duplicates: []string{}, Rejected: append([]string{}, rejected...)}
	if len(accepted) == 0 {
		return result, nil
	}

	s.adding.Lock()
	defer s.adding.Unlock()
	if s.closed {
		return result, ErrClosed
	}

	fresh := make([]string, 0, len(accepted))
	for _, u := range accepted {
		if s.known(u) {
			result.Duplicates = append(result.Duplicates, u)
			continue
		}
		fresh = append(fresh, u)
	}
	if len(fresh) == 0 {
		return result, nil
	}

	eng := s.engine()
	if eng == nil {
		if err := s.start(fresh); err != nil {
			return result, err
		}
		result.Added = fresh
		return result, nil
	}
	for _, u := range fresh {
		if _, ok := eng.AddSource(u); !ok {
			return result, ErrClosed
		}
		result.Added = append(result.Added, u)
	}
	return result, nil
}

// Retry re-queues one failed entry. It is false when there is no such failed entry.
func (s *Session) Retry(entry int) bool {
	eng := s.engine()
	return eng != nil && eng.Retry(entry)
}

// RetryPlaylist re-queues the failed entries of a playlist and says how many.
func (s *Session) RetryPlaylist(playlist int) int {
	eng := s.engine()
	if eng == nil {
		return 0
	}
	return eng.RetryPlaylist(playlist)
}

// known reports whether the run already has this URL and it did not fail.
func (s *Session) known(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, pl := range s.state.Playlists {
		if pl.URL == url && pl.Err == "" {
			return true
		}
	}
	return false
}
