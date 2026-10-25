package preferencessession

import "youtube-downloader/libs/mvd-core/settings"

// folderOfItem is the folder the setting of the cursor holds now.
func (s Session) folderOfItem() string {
	if s.Cursor == itemOutputDir {
		return s.Cfg.OutputDir
	}

	return s.Cfg.LogDir
}

// ChooseFolder asks for a folder for the setting of the cursor: the host's own chooser when it has
// one, else its folder browser.
func (s *Session) ChooseFolder() Effect {
	start := s.folderOfItem()
	if !s.CanPickFolder {
		s.Mode = settings.EditFolder

		return Effect{Kind: BrowseFolders, Folder: start}
	}
	s.Mode = settings.EditPicking

	return Effect{Kind: PickFolder, Folder: start, Item: s.Cursor}
}

// FolderPicked takes the answer of the host's chooser for setting item. When the chooser could not
// be shown, the folder browser is asked for instead, so the setting can still be changed.
func (s *Session) FolderPicked(item int, path string, chosen bool, err error) Effect {
	s.Mode = settings.EditNone
	switch {
	case err != nil:
		s.Mode = settings.EditFolder

		return Effect{Kind: BrowseFolders, Folder: s.folderOfItem()}
	case chosen:
		if item == itemOutputDir {
			s.Cfg.OutputDir = path
		} else {
			s.Cfg.LogDir = path
		}
	}

	return Effect{}
}

// FolderChosen takes the folder the browser ended on for the setting of the cursor.
func (s *Session) FolderChosen(dir string) {
	if s.Cursor == itemOutputDir {
		s.Cfg.OutputDir = dir
	} else {
		s.Cfg.LogDir = dir
	}
	s.Mode = settings.EditNone
}

// FolderCancelled ends the browsing with nothing changed.
func (s *Session) FolderCancelled() { s.Mode = settings.EditNone }
