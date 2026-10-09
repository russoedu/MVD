package main

import (
	"fmt"
	"testing"
)

// songs makes a baseline of n songs and their answers now: the ones from index differentFrom
// on answer with no video.
func songs(n int, differentFrom int) (Baseline, []Outcome) {
	var baseline Baseline
	var outcomes []Outcome
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("s%02d", i)
		baseline.Songs = append(baseline.Songs, Song{ID: id, Official: "v" + id})
		found := "v" + id
		if i >= differentFrom {
			found = ""
		}
		outcomes = append(outcomes, Outcome{ID: id, Expected: "v" + id, Found: found, Known: true})
	}

	return baseline, outcomes
}

func TestAFewDifferentSongsAreNotAChangeOfYouTube(t *testing.T) {
	baseline, outcomes := songs(50, 47) // 3 of 50 differ

	verdict := Judge(baseline, outcomes, 0.1)

	if verdict.Failed || len(verdict.Different) != 3 {
		t.Errorf("3 of 50 should pass and be listed: %+v", verdict)
	}
}

func TestManyDifferentSongsAreAChangeOfYouTube(t *testing.T) {
	baseline, outcomes := songs(50, 40) // 10 of 50 differ

	verdict := Judge(baseline, outcomes, 0.1)

	if !verdict.Failed || verdict.Reason == "" {
		t.Errorf("10 of 50 should fail with a reason: %+v", verdict)
	}
}

func TestTooFewKnownSongsCannotTell(t *testing.T) {
	baseline, outcomes := songs(10, 10)

	if verdict := Judge(baseline, outcomes, 0.1); !verdict.Failed {
		t.Errorf("10 known songs are not enough to say nothing changed: %+v", verdict)
	}
}

func TestNewAndGoneSongsAreCountedButDoNotDecide(t *testing.T) {
	baseline, outcomes := songs(50, 50)
	baseline.Songs = append(baseline.Songs, Song{ID: "removed"})
	outcomes = append(outcomes, Outcome{ID: "fresh", Found: "x"})

	verdict := Judge(baseline, outcomes, 0.1)

	if verdict.Failed || verdict.Gone != 1 || len(verdict.New) != 1 || verdict.Known != 50 {
		t.Errorf("verdict: %+v", verdict)
	}
}
