package songfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadTextAndCSV(t *testing.T) {
	list, err := Load(write(t, "90s dance.txt", "ATB - Killer\nSnap! - Rhythm Is a Dancer\nbad line\n"))
	if err != nil || list.Title != "90s dance" || len(list.Songs) != 2 || list.Skipped != 1 {
		t.Fatalf("text: got %+v, %v", list, err)
	}

	list, err = Load(write(t, "export.csv", "Title,Artist,Duration (ms)\nKiller,ATB,248000\n"))
	if err != nil || list.Title != "export" || len(list.Songs) != 1 || list.Songs[0].DurationMs != 248000 {
		t.Fatalf("csv: got %+v, %v", list, err)
	}

	// A CSV without the columns is read as lines.
	list, err = Load(write(t, "plain.csv", "ATB - Killer\n"))
	if err != nil || len(list.Songs) != 1 || list.Songs[0].Artist != "ATB" {
		t.Fatalf("csv as lines: got %+v, %v", list, err)
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Error("a missing file is an error")
	}
	_, err := Load(write(t, "empty.txt", "just words\n"))
	if err == nil || !strings.Contains(err.Error(), "Artist - Title") {
		t.Errorf("a file without songs should say how to write them, got %v", err)
	}
}
