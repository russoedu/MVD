package review

import (
	"fmt"
	"path/filepath"
	"strings"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Save asks where to save the review as a .mvd file: in the system's window when the host has one,
// else on the bottom line. A live review is saved with the list, so it does nothing here.
func (s *Session) Save() []Effect {
	if s.live != nil {
		return nil
	}
	if s.opts.CanSaveFile {
		s.Picking = true
		dir, name := filepath.Split(s.defaultSavePath())

		return []Effect{PickSaveFile{Dir: filepath.Clean(dir), Name: name}}
	}
	s.ask(inputSave, "Save as: ", s.defaultSavePath())

	return nil
}

// Load asks which .mvd file to open: in the system's window when the host has one, else on the
// bottom line. A live review does nothing here.
func (s *Session) Load() []Effect {
	if s.live != nil {
		return nil
	}
	if s.opts.CanPickFile {
		s.Picking = true

		return []Effect{PickOpenFile{Start: s.opts.SaveDir}}
	}
	s.ask(inputLoad, "Open the file: ", s.opts.SaveDir)

	return nil
}

// FilePicked is the host's answer to PickOpenFile or PickSaveFile (save says which): chosen is
// false when the person cancelled, and err when the window could not be shown, in which case the
// name is typed instead.
func (s *Session) FilePicked(path string, chosen bool, err error, save bool) {
	s.Picking = false
	switch {
	case err != nil && save:
		s.ask(inputSave, "Save as: ", s.defaultSavePath())
	case err != nil:
		s.ask(inputLoad, "Open the file: ", s.opts.SaveDir)
	case chosen && save:
		s.saveFile(path)
	case chosen:
		s.loadFile(path)
	}
}

// saveFile writes the review to a .mvd file.
func (s *Session) saveFile(path string) {
	if filepath.Ext(path) == "" {
		path += playlistfile.Extension
	}
	if err := playlistfile.Save(path, playlistfile.FromSongs(s.Rows)); err != nil {
		s.Flash("cannot save: " + err.Error())

		return
	}
	s.Flash("saved " + path)
}

// loadFile replaces the review with the songs of a saved .mvd file.
func (s *Session) loadFile(path string) {
	file, err := playlistfile.Load(path)
	if err != nil {
		s.Flash("cannot open it: " + err.Error())

		return
	}
	s.Rows = file.Songs()
	s.Selected = 0
	s.autosave()
	s.Flash(fmt.Sprintf("opened %d songs from %s", len(s.Rows), filepath.Base(path)))
}

// defaultSavePath is the file offered when saving: the first playlist's name, in the folder for
// saved plans.
func (s Session) defaultSavePath() string {
	name := "playlist"
	if len(s.Rows) > 0 && s.Rows[0].Playlist != "" {
		name = s.Rows[0].Playlist
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 32, strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}

	return filepath.Join(s.opts.SaveDir, strings.TrimSpace(b.String())+playlistfile.Extension)
}
