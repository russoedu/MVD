package folderbrowser

import "path/filepath"

// Browser is a directory navigator: browse into and out of folders, make a new one. It is a value:
// a copy keeps its own state, and Entries is replaced, never changed in place.
type Browser struct {
	// Dir is the folder being browsed.
	Dir string
	// Entries are the names of its sub-folders.
	Entries []string
	// Cursor is the index in Entries of the sub-folder the person is on.
	Cursor int
	// Creating is true while the name of a new folder is being asked; NewName is what was typed.
	Creating bool
	NewName  string
	// Err is what went wrong the last time the disk was asked, "" when nothing did.
	Err  string
	disk Disk
}

// New is a browser that begins in start (see ResolveDir) and reads the given disk, the
// operating system's when disk is nil.
func New(start string, disk Disk) Browser {
	if disk == nil {
		disk = OSDisk{}
	}
	b := Browser{Dir: ResolveDir(start), disk: disk}
	b.load()

	return b
}

// Listing is a browser showing a listing that was not read from disk: for previews and tests,
// where the folders must not depend on the machine.
func (b Browser) Listing(dir string, entries []string, cursor int, err string, creating bool) Browser {
	b.Dir, b.Entries, b.Cursor, b.Err, b.Creating = dir, entries, cursor, err, creating

	return b
}

// load reads the sub-folders of Dir and puts the cursor on the first.
func (b *Browser) load() {
	b.Entries = nil
	b.Cursor = 0
	b.Err = ""
	names, err := b.disk.Subfolders(b.Dir)
	if err != nil {
		b.Err = err.Error()

		return
	}
	b.Entries = names
}

// Up moves the cursor to the sub-folder before.
func (b *Browser) Up() {
	if b.Cursor > 0 {
		b.Cursor--
	}
}

// Down moves the cursor to the sub-folder after.
func (b *Browser) Down() {
	if b.Cursor < len(b.Entries)-1 {
		b.Cursor++
	}
}

// Open goes into the sub-folder the cursor is on.
func (b *Browser) Open() {
	if b.Cursor < len(b.Entries) {
		b.Dir = filepath.Join(b.Dir, b.Entries[b.Cursor])
		b.load()
	}
}

// Parent goes to the folder that holds this one, when there is one.
func (b *Browser) Parent() {
	if parent := filepath.Dir(b.Dir); parent != b.Dir {
		b.Dir = parent
		b.load()
	}
}
