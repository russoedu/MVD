package openfolder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEachSystemHasItsOwnFileManager(t *testing.T) {
	for goos, want := range map[string]string{"windows": "explorer.exe", "darwin": "open", "linux": "xdg-open", "freebsd": "xdg-open"} {
		name, args := command(goos, `/some folder/with spaces`)
		if name != want || len(args) != 1 || args[0] != `/some folder/with spaces` {
			t.Errorf("%s: %s %v, want %s and the path as one argument", goos, name, args, want)
		}
	}
}

func TestAFolderThatIsNotThereCannotBeOpened(t *testing.T) {
	if err := Open(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("a folder that does not exist was opened")
	}
}

func TestAFileIsNotAFolder(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Open(file); err == nil {
		t.Error("a file was opened as a folder")
	}
}
