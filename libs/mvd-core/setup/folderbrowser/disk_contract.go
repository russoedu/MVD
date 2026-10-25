package folderbrowser

// Disk is what the browser needs from the machine's folders.
type Disk interface {
	// Subfolders are the names of the folders inside dir, sorted.
	Subfolders(dir string) ([]string, error)
	// MakeFolder makes the folder at path.
	MakeFolder(path string) error
}
