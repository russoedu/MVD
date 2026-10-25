package runcontrol

import (
	"testing"

	"youtube-downloader/libs/mvd-core/runstate"
)

func TestARunIsFinishedOnlyWhenEverythingIsReadIdleAndNothingWaitsForTheChoice(t *testing.T) {
	done := Facts{HasRun: true, Idle: true, AllListed: true}
	if !done.Finished() || done.Active() {
		t.Error("an idle run with everything read is finished")
	}
	waiting := done
	waiting.Tally = runstate.Tally{Planned: 2}
	if waiting.Finished() || !waiting.Active() {
		t.Error("a run with songs waiting for the choice is not finished")
	}
	reading := done
	reading.AllListed = false
	if reading.Finished() {
		t.Error("a run still reading playlists is not finished")
	}
	stopping := done
	stopping.Stopping = true
	if stopping.Finished() {
		t.Error("a run being wound down is not finished")
	}
}

func TestTheListCanBeEditedOnlyWithoutARun(t *testing.T) {
	if !(Facts{}).Editable() {
		t.Error("with no run the list can be edited")
	}
	if (Facts{HasRun: true}).Editable() || (Facts{Starting: true}).Editable() || (Facts{Stopping: true}).Editable() {
		t.Error("a run, or one starting or stopping, locks the list")
	}
}

func TestARunThatHoldsOnlyTheListIsOneThatIsStoppedOrHasNothingComingIn(t *testing.T) {
	cases := []struct {
		name string
		f    Facts
		want bool
	}{
		{"no run", Facts{}, false},
		{"nothing coming in", Facts{HasRun: true}, true},
		{"songs coming in", Facts{HasRun: true, Tally: runstate.Tally{Running: 3}}, false},
		{"stopped with songs on their way", Facts{HasRun: true, Paused: true, Tally: runstate.Tally{Running: 8}}, true},
		{"winding down", Facts{HasRun: true, Stopping: true}, false},
	}
	for _, c := range cases {
		if got := c.f.HoldsOnlyTheList(); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestLinksCanBeAddedOnlyToARunThatTakesThemAndIsActive(t *testing.T) {
	f := Facts{HasRun: true, CanAddLinks: true}
	if !f.Addable() {
		t.Error("an active run that takes links is addable")
	}
	f.CanAddLinks = false
	if f.Addable() {
		t.Error("a run that cannot take links is not addable")
	}
	f = Facts{HasRun: true, CanAddLinks: true, Idle: true, AllListed: true}
	if f.Addable() {
		t.Error("a finished run is settled, not added to")
	}
}

func TestClosingLosesOnlyWhatIsComingInOrAccepted(t *testing.T) {
	tally := runstate.Tally{Running: 2, Accepted: 3}
	if got := (Facts{HasRun: true, Tally: tally}).Downloading(); got != 5 {
		t.Errorf("a run that goes on would lose %d", got)
	}
	if got := (Facts{HasRun: true, Paused: true, Tally: tally}).Downloading(); got != 2 {
		t.Errorf("a stopped run would lose only what is on its way, got %d", got)
	}
	if got := (Facts{Tally: tally}).Downloading(); got != 0 {
		t.Errorf("no run loses nothing, got %d", got)
	}
}
