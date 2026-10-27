package runcontrol

// Lifecycle is where a run is between being asked for and being gone: it is being started, or
// wound down, and what the person asked for meanwhile waits until it has been.
type Lifecycle struct {
	// Starting and Stopping say a run is being started or wound down.
	Starting, Stopping bool
	// After is what the person asked for while a run only held the list: it is done once the run has
	// been stopped. Nil for nothing.
	After any
}

// Busy says a run is being started or wound down, so nothing else can be done yet.
func (l Lifecycle) Busy() bool { return l.Starting || l.Stopping }

// BeginStart notes that a run is being started.
func (l *Lifecycle) BeginStart() { l.Starting = true }

// Started notes that the start has ended, whether or not it worked.
func (l *Lifecycle) Started() { l.Starting = false }

// BeginStop notes that the run is being wound down.
func (l *Lifecycle) BeginStop() { l.Stopping = true }

// WaitThen notes what to do once the run has been wound down.
func (l *Lifecycle) WaitThen(next any) { l.After = next }

// Closed notes that the run has been wound down, and returns what was waiting for that, once.
func (l *Lifecycle) Closed() any {
	l.Stopping = false
	next := l.After
	l.After = nil

	return next
}
