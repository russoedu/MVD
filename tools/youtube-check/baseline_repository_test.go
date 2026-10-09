package main

import (
	"path/filepath"
	"testing"
)

func TestABaselineIsSavedSortedAndReadBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	in := Baseline{Playlist: "p", Songs: []Song{{ID: "b", Official: "x"}, {ID: "a"}}}

	if err := saveBaseline(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := loadBaseline(path)
	if err != nil {
		t.Fatal(err)
	}

	if out.Playlist != "p" || len(out.Songs) != 2 || out.Songs[0].ID != "a" || out.Songs[1].Official != "x" {
		t.Errorf("baseline: %+v", out)
	}
}

func TestAMissingBaselineIsEmpty(t *testing.T) {
	baseline, err := loadBaseline(filepath.Join(t.TempDir(), "none.json"))
	if err != nil || len(baseline.Songs) != 0 {
		t.Errorf("a missing baseline should be empty: %+v, %v", baseline, err)
	}
}
