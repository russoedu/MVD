package session

// Option changes how a Session behaves. Pass them to New.
type Option func(*Session)

// WithFailureHook calls hook each time an entry fails to download or a link fails to be
// looked up, from the goroutine that folds the engine's events into the state, with no
// lock held.
//
// The hook must return quickly: until it does, no further event is applied and the
// page stops updating. A hook that has work to do should start it on its own goroutine.
func WithFailureHook(hook func()) Option {
	return func(s *Session) { s.onFailure = hook }
}
