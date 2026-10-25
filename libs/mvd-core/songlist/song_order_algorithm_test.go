package songlist

import (
	"testing"

	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runstate"
)

func entry(index int, title string) *runstate.Entry {
	return &runstate.Entry{EntryInfo: engine.EntryInfo{Index: index, Title: title, VideoID: "v"}}
}

func TestTheSongsKeepTheirOrderUntilAHeadingIsClicked(t *testing.T) {
	entries := []*runstate.Entry{entry(1, "b"), entry(2, "a")}
	got := Order{}.Sorted([]int{0, 1}, entries)
	if got[0] != 0 || got[1] != 1 {
		t.Fatalf("an unsorted list changed order: %v", got)
	}
}

func TestAClickSortsAscendingThenTurnsTheOrderRound(t *testing.T) {
	entries := []*runstate.Entry{entry(1, "b"), entry(2, "A"), entry(3, "c")}

	o := Order{}.Clicked(ColumnSong)
	if o.By != ColumnSong || o.Desc {
		t.Fatalf("first click gave %+v", o)
	}
	if got := o.Sorted([]int{0, 1, 2}, entries); got[0] != 1 || got[2] != 2 {
		t.Errorf("ascending, ignoring case: %v", got)
	}

	o = o.Clicked(ColumnSong)
	if !o.Desc {
		t.Fatalf("second click gave %+v", o)
	}
	if got := o.Sorted([]int{0, 1, 2}, entries); got[0] != 2 || got[2] != 1 {
		t.Errorf("descending: %v", got)
	}
}

func TestTheNumberColumnSortsFirstWhenNothingSortsYet(t *testing.T) {
	if d := (Order{}).Direction(ColumnIdx); d != 1 {
		t.Errorf("the # column of an unsorted list says %d, want ascending", d)
	}
	if d := (Order{}).Direction(ColumnSong); d != 0 {
		t.Errorf("the song column of an unsorted list says %d, want none", d)
	}
	if o := (Order{}).Clicked(ColumnIdx); !o.Desc {
		t.Errorf("a click on the # column should turn it round, got %+v", o)
	}
}

func TestSortingDoesNotChangeTheListItWasGiven(t *testing.T) {
	entries := []*runstate.Entry{entry(1, "b"), entry(2, "a")}
	ids := []int{0, 1}
	Order{By: ColumnSong}.Sorted(ids, entries)
	if ids[0] != 0 || ids[1] != 1 {
		t.Errorf("the ids were sorted in place: %v", ids)
	}
}
