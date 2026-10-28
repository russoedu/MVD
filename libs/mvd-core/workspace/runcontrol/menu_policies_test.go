package runcontrol

import "testing"

func TestRunMenuOfWithoutRun(t *testing.T) {
	m := RunMenuOf(Facts{}, false, false)
	if !m.Start || m.Pause || m.Resume || m.Stop || m.Retry || m.DownloadAll {
		t.Fatalf("a list with no run can only start: %+v", m)
	}
	if RunMenuOf(Facts{}, true, false).Start || RunMenuOf(Facts{}, false, true).Start {
		t.Fatal("an empty list, or tools that are not ready, cannot start")
	}
}

func TestRunMenuOfPausedAndBusy(t *testing.T) {
	paused := RunMenuOf(Facts{HasRun: true, Paused: true}, false, false)
	if !paused.Resume || paused.Pause || !paused.Stop || paused.Retry {
		t.Fatalf("a paused run resumes, and has nothing to retry: %+v", paused)
	}
	failed := Facts{HasRun: true, Paused: true}
	failed.Tally.Failed = 1
	if !RunMenuOf(failed, false, false).Retry {
		t.Fatal("a run with a failed song can retry it")
	}
	busy := RunMenuOf(Facts{HasRun: true, Stopping: true}, false, false)
	if busy.Stop || busy.Retry || busy.Start {
		t.Fatalf("nothing can be done while the run winds down: %+v", busy)
	}
}

func TestRefusalAndReason(t *testing.T) {
	if Refusal(Facts{HasRun: true}, 1, true) != "The downloads are already going" {
		t.Fatal("a run is going")
	}
	if Refusal(Facts{}, 0, true) == "" || Refusal(Facts{}, 1, false) == "" || Refusal(Facts{}, 1, true) != "" {
		t.Fatal("refusals")
	}
	if DisabledReason(Facts{Starting: true}, true, "tools") != "One moment..." ||
		DisabledReason(Facts{}, true, "tools") != "tools" ||
		DisabledReason(Facts{}, false, "") != "Add a source first." {
		t.Fatal("reasons")
	}
}

func TestEditMenuOf(t *testing.T) {
	free := EditMenuOf(Facts{}, EditFacts{HasCursor: true, Selected: 1, Count: 3, CanUndo: true})
	if !free.Paste || !free.Delete || !free.Undo || free.Redo || !free.MoveUp || !free.MoveDown || free.Cut {
		t.Fatalf("a free list: %+v", free)
	}
	writing := EditMenuOf(Facts{}, EditFacts{WritingAddress: true, HasSelection: true, HasCursor: true, Count: 2})
	if !writing.Cut || !writing.Copy || writing.Delete || writing.MoveDown {
		t.Fatalf("writing an address: %+v", writing)
	}
	live := EditMenuOf(Facts{HasRun: true, CanAddLinks: true, Idle: false}, EditFacts{PlaylistSelected: true})
	if !live.Delete || !live.Paste || live.Undo {
		t.Fatalf("a run that takes links: %+v", live)
	}
}
