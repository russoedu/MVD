package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceFilePutsTheNewFileInPlaceAndLeavesNothingElse(t *testing.T) {
	dir := t.TempDir()
	final := filepath.Join(dir, "tool")
	partial := final + ".part"
	if err := os.WriteFile(final, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(partial, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(partial, final); err != nil {
		t.Fatal(err)
	}

	if read(t, final) != "new" {
		t.Error("the file was not replaced")
	}
	if got := listing(t, dir); got != "tool" {
		t.Errorf("folder holds %q, want only the tool", got)
	}
}

func TestReplaceFileInstallsWhenThereIsNothingToReplace(t *testing.T) {
	dir := t.TempDir()
	final := filepath.Join(dir, "tool")
	partial := final + ".part"
	if err := os.WriteFile(partial, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(partial, final); err != nil {
		t.Fatal(err)
	}

	if read(t, final) != "new" {
		t.Error("the file was not installed")
	}
}

func TestReplaceFilePutsTheOldFileBackWhenTheNewOneCannotBePlaced(t *testing.T) {
	dir := t.TempDir()
	final := filepath.Join(dir, "tool")
	if err := os.WriteFile(final, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := replaceFile(filepath.Join(dir, "does-not-exist.part"), final)

	if err == nil {
		t.Fatal("expected an error")
	}
	if read(t, final) != "old" {
		t.Error("the working copy was lost when the replacement failed")
	}
	if strings.Contains(listing(t, dir), ".old") {
		t.Errorf("a .old file was left: %s", listing(t, dir))
	}
}
