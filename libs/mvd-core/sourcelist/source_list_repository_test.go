package sourcelist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "list.txt")

	if urls, err := Load(path); err != nil || urls != nil {
		t.Fatalf("missing file should be empty: %v %v", urls, err)
	}

	if err := Save(path, []string{"https://a", "  ", "https://b", ""}); err != nil {
		t.Fatal(err)
	}
	urls, err := Load(path)
	if err != nil || len(urls) != 2 || urls[0] != "https://a" || urls[1] != "https://b" {
		t.Fatalf("round-trip wrong: %v %v", urls, err)
	}

	// Comments and blanks are ignored on load.
	os.WriteFile(path, []byte("# header\n\nhttps://c\n"), 0o644)
	if urls, _ := Load(path); len(urls) != 1 || urls[0] != "https://c" {
		t.Fatalf("comment handling wrong: %v", urls)
	}

	if err := Clear(path); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Error("clear should remove the file")
	}
	if err := Clear(path); err != nil {
		t.Error("clearing a missing file should be fine")
	}
}
