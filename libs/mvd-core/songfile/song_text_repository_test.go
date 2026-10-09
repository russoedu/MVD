package songfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var noon = time.Date(2026, 10, 9, 12, 30, 0, 0, time.UTC)

func TestSaveNamesTheListAfterItsFirstComment(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "songs")
	path, err := Save(dir, "# Road: trip/2026?\r\nATB - Killer\r\nSeal - Kiss From a Rose\r\n", noon)
	if err != nil || filepath.Base(path) != "Road trip 2026.txt" {
		t.Fatalf("got %q, %v", path, err)
	}
	list, err := Load(path)
	if err != nil || len(list.Songs) != 2 || list.Title != "Road trip 2026" {
		t.Fatalf("saved list reads back as %+v, %v", list, err)
	}

	// The same name again does not overwrite.
	again, err := Save(dir, "# Road trip 2026\nPrince - Kiss\n", noon)
	if err != nil || filepath.Base(again) != "Road trip 2026 (2).txt" {
		t.Fatalf("got %q, %v; want a numbered copy", again, err)
	}
}

func TestSaveWithoutAComment(t *testing.T) {
	path, err := Save(t.TempDir(), "ATB - Killer", noon)
	if err != nil || filepath.Base(path) != "Songs 2026-10-09 1230.txt" {
		t.Fatalf("got %q, %v", path, err)
	}
}

func TestSaveKeepsACSVAsCSV(t *testing.T) {
	path, err := Save(t.TempDir(), "Title,Artist\nKiller,ATB\n", noon)
	if err != nil || filepath.Ext(path) != ".csv" {
		t.Fatalf("got %q, %v", path, err)
	}
	if list, err := Load(path); err != nil || len(list.Songs) != 1 || list.Songs[0].Artist != "ATB" {
		t.Fatalf("got %+v, %v", list, err)
	}
}

func TestSaveRefusesATextWithoutSongs(t *testing.T) {
	dir := t.TempDir()
	if _, err := Save(dir, "# only a name\nnothing here\n", noon); err == nil || !strings.Contains(err.Error(), "Artist - Title") {
		t.Errorf("got %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("nothing should be stored, found %d files", len(entries))
	}
}
