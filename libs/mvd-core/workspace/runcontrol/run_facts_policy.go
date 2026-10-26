package runcontrol

import "youtube-downloader/libs/mvd-core/runstate"

// Facts are what the workspace knows of the run at a moment.
type Facts struct {
	// HasRun says a run is attached; Paused that its downloads are held.
	HasRun, Paused bool
	// Starting and Stopping say a run is being started or wound down.
	Starting, Stopping bool
	// CanAddLinks says the run takes more links while it works.
	CanAddLinks bool
	// Idle and AllListed are the run's own report: nothing is queued or busy, every playlist was read.
	Idle, AllListed bool
	// Tally is the run's counters.
	Tally runstate.Tally
}

// Busy says a run is being started or wound down, so nothing else can be done yet.
func (f Facts) Busy() bool { return f.Starting || f.Stopping }

// Finished says the run has nothing more to do: every playlist was read, the queue is empty with no
// worker busy, and no song is left waiting for the person to choose it. A song that is not wanted
// is dealt with by stopping the run.
func (f Facts) Finished() bool {
	if !f.HasRun || f.Busy() {
		return false
	}

	return f.Idle && f.AllListed && f.Tally.Planned == 0
}

// Active says a run is going on that has to be stopped before the list can change: one that has
// finished does not.
func (f Facts) Active() bool { return f.HasRun && !f.Finished() }

// Editable says the list can be changed now: not while a run has its own copy.
func (f Facts) Editable() bool { return !f.HasRun && !f.Busy() }

// Addable says links can be added to the run that is going on.
func (f Facts) Addable() bool { return f.CanAddLinks && f.Active() && !f.Busy() }

// Waiting says some song of the run waits for the person to choose it.
func (f Facts) Waiting() bool { return f.HasRun && f.Tally.Planned > 0 }

// HoldsOnlyTheList says the run does nothing that stopping it would lose: it is stopped, or no song
// is coming in. Such a run is only the list loaded, and what the person asks for next does not have
// to wait for them to stop it.
func (f Facts) HoldsOnlyTheList() bool {
	if !f.HasRun || f.Busy() {
		return false
	}

	return f.Paused || f.Tally.Running == 0
}

// Downloading is how many songs are being downloaded or wait for their turn to be.
func (f Facts) Downloading() int {
	if !f.HasRun {
		return 0
	}
	if f.Paused {
		return f.Tally.Running // what was accepted waits for the Download button: nothing is lost by quitting
	}

	return f.Tally.Accepted + f.Tally.Running
}
