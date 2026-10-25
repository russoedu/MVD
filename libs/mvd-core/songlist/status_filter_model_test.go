package songlist

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func inState(s engine.EntryState) *runstate.Entry {
	return &runstate.Entry{State: s}
}

func TestEveryEngineStateBelongsToAGroupTheFilterListsOrToNone(t *testing.T) {
	listed := map[string]bool{}
	for _, g := range Groups {
		listed[g.Key] = true
	}
	for _, s := range []engine.EntryState{engine.StatePlanned, engine.StateAccepted, engine.StateQueued, engine.StateResolving, engine.StateDownloading, engine.StateMerging, engine.StateDone, engine.StateDuplicate, engine.StateNotFound, engine.StateSkipped, engine.StateFailed} {
		if g := GroupOf(inState(s)); !listed[g] {
			t.Errorf("state %v is in the group %q, which the filter does not list", s, g)
		}
	}
}

func TestAPlannedSongIsOfficialOnlyWhenAnOfficialVideoWasFound(t *testing.T) {
	plain := inState(engine.StatePlanned)
	if g := GroupOf(plain); g != GroupOriginal {
		t.Errorf("a planned song with nothing found is %q", g)
	}
	own := &runstate.Entry{State: engine.StatePlanned, OwnOfficial: true}
	if g := GroupOf(own); g != GroupOfficial {
		t.Errorf("a planned song whose own upload is official is %q", g)
	}
}

func TestFilterHidesAndShowsGroupsWithoutChangingTheCopyItCameFrom(t *testing.T) {
	none := Filter{}
	if none.Active() || !none.Matches(inState(engine.StateFailed)) {
		t.Fatal("an empty filter should show everything")
	}

	hidden := none.Toggled(GroupFailed)
	if !hidden.Active() || hidden.Matches(inState(engine.StateFailed)) || !hidden.Matches(inState(engine.StateDone)) {
		t.Errorf("hiding the failed songs: active %v", hidden.Active())
	}
	if none.Active() {
		t.Error("toggling changed the filter it was called on")
	}
	if hidden.Toggled(GroupFailed).Active() {
		t.Error("toggling a hidden group again should show it")
	}
}

func TestOnlyShowsOneGroupAndShowsAllWhenAskedAgain(t *testing.T) {
	only := Filter{}.Only(GroupFailed)
	if only.HiddenCount() != len(Groups)-1 || only.Hidden(GroupFailed) {
		t.Fatalf("only failed: %d hidden", only.HiddenCount())
	}
	if again := only.Only(GroupFailed); again.Active() {
		t.Error("asking for the same group again should show them all")
	}
}

func TestTheStatusWordCarriesATone(t *testing.T) {
	word, tone := StatusWord(inState(engine.StateFailed))
	if word != "failed" || tone != ToneError {
		t.Errorf("a failed song says %q in tone %v", word, tone)
	}
	word, tone = StatusWord(&runstate.Entry{State: engine.StateDownloading, Total: 10})
	if word != "downloading" || tone != ToneHighlight {
		t.Errorf("a song coming in says %q in tone %v", word, tone)
	}
	if word, _ = StatusWord(inState(engine.StateDownloading)); word != "starting" {
		t.Errorf("a song with no size yet says %q", word)
	}
}
