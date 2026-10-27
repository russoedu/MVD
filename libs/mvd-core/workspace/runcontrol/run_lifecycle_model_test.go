package runcontrol

import "testing"

func TestLifecycleGivesBackWhatWaitedForTheStopOnce(t *testing.T) {
	var l Lifecycle
	l.WaitThen("open")
	l.BeginStop()
	if !l.Busy() {
		t.Fatal("a run being stopped is busy")
	}
	if got := l.Closed(); got != "open" {
		t.Fatalf("got %v, want open", got)
	}
	if l.Busy() || l.Closed() != nil {
		t.Fatal("what waited is given back once, and the run is no longer busy")
	}
	l.BeginStart()
	if !l.Busy() {
		t.Fatal("a run being started is busy")
	}
	l.Started()
	if l.Busy() {
		t.Fatal("a started run is not busy")
	}
}
