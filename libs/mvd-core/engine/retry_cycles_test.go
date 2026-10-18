package engine

import "testing"

func TestSweepCandidatesCountTheCycle(t *testing.T) {
	e := &Engine{opts: Options{AutoRetry: true, RetryCycles: 3}}
	e.entries = []*engineEntry{
		{info: EntryInfo{ID: 0}, state: StateFailed, deferred: true, sweeps: 1},
		{info: EntryInfo{ID: 1}, state: StateFailed, deferred: true, sweeps: 0},
		{info: EntryInfo{ID: 2}, state: StateFailed, deferred: true, sweeps: 3}, // out of cycles
		{info: EntryInfo{ID: 3}, state: StateFailed},                            // permanent
		{info: EntryInfo{ID: 4}, state: StateDone, deferred: true},
	}
	ids, cycle := e.sweepCandidates()
	if len(ids) != 2 || ids[0] != 0 || ids[1] != 1 {
		t.Errorf("candidates = %v, want [0 1]", ids)
	}
	if cycle != 1 {
		t.Errorf("cycle = %d, want 1: the wait follows the entry with the fewest retries", cycle)
	}

	e.entries[1].sweeps = 2
	if _, cycle = e.sweepCandidates(); cycle != 2 {
		t.Errorf("cycle = %d, want 2", cycle)
	}
}

func TestRetryCyclesDefaultToThirty(t *testing.T) {
	if got := (&Engine{}).retryCycles(); got != 30 {
		t.Errorf("retryCycles = %d, want 30", got)
	}
}
