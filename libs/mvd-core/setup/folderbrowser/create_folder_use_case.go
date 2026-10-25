package folderbrowser

import (
	"path/filepath"
	"strings"
)

// StartCreating begins asking for the name of a new folder.
func (b *Browser) StartCreating() {
	b.Creating = true
	b.NewName = ""
}

// SetNewName is the name typed so far, as the renderer's text box holds it.
func (b *Browser) SetNewName(name string) { b.NewName = name }

// CancelCreating stops asking for the name.
func (b *Browser) CancelCreating() { b.Creating = false }

// CreateFolder makes the folder that was named inside Dir and goes into it. A blank name makes
// nothing; a folder the disk refuses leaves the reason in Err.
func (b *Browser) CreateFolder() {
	name := strings.TrimSpace(b.NewName)
	if name != "" {
		path := filepath.Join(b.Dir, name)
		if err := b.disk.MakeFolder(path); err != nil {
			b.Err = err.Error()
		} else {
			b.Dir = path
			b.load()
		}
	}
	b.Creating = false
}
