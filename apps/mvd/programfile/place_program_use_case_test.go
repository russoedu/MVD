package programfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlaceCopiesTheBytesIntoAFolderItCreatesAndLeavesNothingElse(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.exe")
	if err := os.WriteFile(src, []byte("program"), 0o755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "new", "folder", "mvd.exe")

	if err := Place(src, dst); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(dst); string(data) != "program" {
		t.Errorf("content = %q", data)
	}
	entries, _ := os.ReadDir(filepath.Dir(dst))
	if len(entries) != 1 {
		t.Errorf("the folder holds %d items, want only the program", len(entries))
	}
}

func TestPlaceReplacesAnOlderCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.exe")
	dst := filepath.Join(dir, "mvd.exe")
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Place(src, dst); err != nil {
		t.Fatal(err)
	}

	if data, _ := os.ReadFile(dst); string(data) != "new" {
		t.Errorf("content = %q", data)
	}
}

func TestPlaceFailsCleanlyWhenThereIsNothingToCopy(t *testing.T) {
	dir := t.TempDir()

	if err := Place(filepath.Join(dir, "missing.exe"), filepath.Join(dir, "out", "mvd.exe")); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(filepath.Join(dir, "out", "mvd.exe.part")); err == nil {
		t.Error("a partial file was left behind")
	}
}
