package workspace

import "youtube-downloader/libs/mvd-core/menubar"

// Names of the entries the workspace acts on.
const (
	entryNew         = "New"
	entryOpen        = "Open…"
	entrySave        = "Save"
	entrySaveAs      = "Save as…"
	entryPreferences = "Preferences…"
	entryQuit        = "Quit"
	entryUndo        = "Undo"
	entryRedo        = "Redo"
	entryCut         = "Cut"
	entryCopy        = "Copy"
	entryPaste       = "Paste"
	entryDelete      = "Delete"
	entryMoveUp      = "Move up"
	entryMoveDown    = "Move down"
	entryAbout       = "About MVD"
	entryOpenFolder  = "Open videos folder"
	entryShortcuts   = "Keyboard shortcuts"
	entryDocs        = "Documentation"
	entryReport      = "Report a problem…"
	entryUpdates     = "Check for updates…"
)

func menus() []menubar.Menu {
	return []menubar.Menu{
		{Name: "File", Entries: []menubar.Entry{
			{Name: entryNew, Shortcut: "Ctrl+N"},
			{Name: entryOpen, Shortcut: "Ctrl+O"},
			{Name: entrySave, Shortcut: "Ctrl+S"},
			{Name: entrySaveAs, Shortcut: "Ctrl+Shift+S"},
			{Separator: true},
			{Name: entryPreferences, Shortcut: "Ctrl+,"},
			{Separator: true},
			{Name: entryQuit, Shortcut: "Ctrl+Q"},
		}},
		{Name: "Edit", Entries: []menubar.Entry{
			{Name: entryUndo, Shortcut: "Ctrl+Z"},
			{Name: entryRedo, Shortcut: "Ctrl+Y"},
			{Separator: true},
			{Name: entryCut, Shortcut: "Ctrl+X"},
			{Name: entryCopy, Shortcut: "Ctrl+C"},
			{Name: entryPaste, Shortcut: "Ctrl+V"},
			{Name: entryDelete, Shortcut: "Del"},
			{Separator: true},
			{Name: entryMoveUp, Shortcut: "Shift+Up"},
			{Name: entryMoveDown, Shortcut: "Shift+Down"},
		}},
		{Name: "Run", Entries: []menubar.Entry{
			{Name: entryStart, Shortcut: "Ctrl+R"},
			{Separator: true},
			{Name: entryPause, Shortcut: "Ctrl+P"},
			{Name: entryResume, Shortcut: "Ctrl+P"},
			{Name: entryStop, Shortcut: "Ctrl+."},
			{Name: entryRetry, Shortcut: "Ctrl+T"},
			{Separator: true},
			{Name: entryAddLink, Shortcut: "Ctrl+L"},
			{Name: entryDownloadAll, Shortcut: "Ctrl+D"},
			{Separator: true},
			{Name: entryOpenFolder, Shortcut: "Ctrl+Shift+O"},
		}},
		{Name: "Help", Entries: []menubar.Entry{
			{Name: entryShortcuts, Shortcut: "F1 / Ctrl+K"},
			{Name: entryDocs},
			{Name: entryReport},
			{Name: entryUpdates},
			{Separator: true},
			{Name: entryAbout},
		}},
	}
}
