package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(d) != "mvd" {
		t.Errorf("app dir should end in mvd, got %q", d)
	}
}

func TestDefaultDownloadsDir(t *testing.T) {
	// Just assert it returns a non-empty absolute-ish path without panicking.
	if DefaultDownloadsDir() == "" {
		t.Error("downloads dir should not be empty")
	}
}

func homeWith(t *testing.T, folders ...string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	for _, f := range folders {
		if err := os.Mkdir(filepath.Join(home, f), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func TestANewInstallPutsItsVideosInAFolderOfItsOwnInVideos(t *testing.T) {
	home := homeWith(t, "Videos", "Downloads")
	if got, want := DefaultVideosDir(), filepath.Join(home, "Videos", "MVD"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOnAMacTheVideosFolderIsCalledMovies(t *testing.T) {
	home := homeWith(t, "Movies", "Downloads")
	if got, want := DefaultVideosDir(), filepath.Join(home, "Movies", "MVD"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWithoutAVideosFolderTheVideosGoToAFolderOfTheirOwnInDownloads(t *testing.T) {
	home := homeWith(t, "Downloads")
	if got, want := DefaultVideosDir(), filepath.Join(home, "Downloads", "MVD"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
