package folderbrowser

import (
	"errors"
	"path/filepath"
	"testing"
)

// fakeDisk is a disk of folders in memory: the keys are folders, the values their sub-folders.
type fakeDisk struct {
	tree map[string][]string
	made []string
	fail error
}

func (d *fakeDisk) Subfolders(dir string) ([]string, error) {
	names, ok := d.tree[dir]
	if !ok {
		return nil, errors.New("open " + dir + ": no such folder")
	}

	return names, nil
}

func (d *fakeDisk) MakeFolder(path string) error {
	if d.fail != nil {
		return d.fail
	}
	d.made = append(d.made, path)
	d.tree[path] = nil

	return nil
}

// musicRoot is the folder the tests begin in, as the browser resolves it on this machine.
func musicRoot() string { return ResolveDir(filepath.Join(string(filepath.Separator), "music")) }

func newDisk() *fakeDisk {
	root := musicRoot()
	return &fakeDisk{tree: map[string][]string{
		root:                         {"90s", "Rock"},
		filepath.Join(root, "90s"):  {"Dance"},
		filepath.Join(root, "Rock"): nil,
	}}
}

func browserAt(t *testing.T, d *fakeDisk) Browser {
	t.Helper()
	b := New(musicRoot(), d)
	if b.Err != "" {
		t.Fatalf("the music folder did not load: %s", b.Err)
	}

	return b
}

func TestTheCursorStaysInsideTheSubfolders(t *testing.T) {
	b := browserAt(t, newDisk())
	b.Up()
	if b.Cursor != 0 {
		t.Errorf("up from the first went to %d", b.Cursor)
	}
	b.Down()
	b.Down()
	if b.Cursor != 1 {
		t.Errorf("down past the last went to %d", b.Cursor)
	}
}

func TestOpeningAFolderAndGoingUpReturnsToTheCursorAtTheTop(t *testing.T) {
	b := browserAt(t, newDisk())
	b.Open()
	if filepath.Base(b.Dir) != "90s" || len(b.Entries) != 1 || b.Cursor != 0 {
		t.Fatalf("opened %q with %v", b.Dir, b.Entries)
	}
	b.Parent()
	if filepath.Base(b.Dir) != "music" || len(b.Entries) != 2 {
		t.Errorf("went up to %q with %v", b.Dir, b.Entries)
	}
}

func TestAFolderThatCannotBeReadShowsWhy(t *testing.T) {
	b := browserAt(t, newDisk())
	b.Dir = filepath.Join(b.Dir, "missing")
	b.load()
	if b.Err == "" || len(b.Entries) != 0 {
		t.Errorf("an unreadable folder should leave an error and no entries: %q %v", b.Err, b.Entries)
	}
}

func TestMakingAFolderGoesIntoIt(t *testing.T) {
	d := newDisk()
	b := browserAt(t, d)
	b.StartCreating()
	b.SetNewName("  Jazz ")
	b.CreateFolder()
	if b.Creating || filepath.Base(b.Dir) != "Jazz" || len(d.made) != 1 {
		t.Errorf("made %v, now in %q, creating %v", d.made, b.Dir, b.Creating)
	}
}

func TestABlankNameMakesNothingAndTheDiskRefusalIsShown(t *testing.T) {
	d := newDisk()
	b := browserAt(t, d)
	b.StartCreating()
	b.SetNewName("   ")
	b.CreateFolder()
	if b.Creating || len(d.made) != 0 {
		t.Errorf("a blank name made %v, creating %v", d.made, b.Creating)
	}

	d.fail = errors.New("access denied")
	b.StartCreating()
	b.SetNewName("Nope")
	before := b.Dir
	b.CreateFolder()
	if b.Err != "access denied" || b.Dir != before || b.Creating {
		t.Errorf("a refused folder: err %q dir %q creating %v", b.Err, b.Dir, b.Creating)
	}
}
