package listhistory

import (
	"testing"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

func list(sources ...string) []playlistfile.Item {
	items := make([]playlistfile.Item, len(sources))
	for i, s := range sources {
		items[i] = playlistfile.Item{Source: s}
	}

	return items
}

func sources(items []playlistfile.Item) string {
	out := ""
	for _, it := range items {
		out += it.Source
	}

	return out
}

func TestAChangeCanBeTakenBackAndMadeAgain(t *testing.T) {
	h := History{}.Recorded(list("a")) // the list was "a", then it became "ab"
	if !h.CanUndo() || h.CanRedo() {
		t.Fatal("after a change it can be taken back, not made again")
	}
	h, before, ok := h.Undone(list("a", "b"))
	if !ok || sources(before) != "a" || h.CanUndo() || !h.CanRedo() {
		t.Fatalf("undone gave %q ok %v", sources(before), ok)
	}
	h, after, ok := h.Redone(list("a"))
	if !ok || sources(after) != "ab" || !h.CanUndo() || h.CanRedo() {
		t.Fatalf("redone gave %q ok %v", sources(after), ok)
	}
}

func TestANewChangeForgetsWhatWasTakenBack(t *testing.T) {
	h := History{}.Recorded(list("a"))
	h, _, _ = h.Undone(list("a", "b"))
	h = h.Recorded(list("a"))
	if h.CanRedo() {
		t.Error("a change after an undo should end the redo")
	}
}

func TestNothingToTakeBackLeavesTheHistoryAsItIs(t *testing.T) {
	h, items, ok := History{}.Undone(list("a"))
	if ok || items != nil || h.CanRedo() {
		t.Errorf("undo with nothing: ok %v items %v", ok, items)
	}
	if _, _, ok := (History{}).Redone(list("a")); ok {
		t.Error("redo with nothing")
	}
}

func TestTheHistoryKeepsCopiesNotTheListItWasGiven(t *testing.T) {
	current := list("a")
	h := History{}.Recorded(current)
	current[0].Source = "changed"
	_, before, _ := h.Undone(list("x"))
	if sources(before) != "a" {
		t.Errorf("the history followed a later change: %q", sources(before))
	}
}
