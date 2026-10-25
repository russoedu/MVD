package folderbrowser

import (
	"os"
	"sort"
)

// OSDisk is the Disk of the operating system.
type OSDisk struct{}

// Subfolders reads the folders inside dir.
func (OSDisk) Subfolders(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	return names, nil
}

// MakeFolder makes the folder at path.
func (OSDisk) MakeFolder(path string) error { return os.Mkdir(path, 0o755) }
