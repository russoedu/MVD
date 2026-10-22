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
	entryCopy        = "Copy"
	entryPaste       = "Paste"
	entryDelete      = "Delete"
	entryMoveUp      = "Move up"
	entryMoveDown    = "Move down"
	entryAbout       = "About MVD"
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
			{Name: entryCopy},
			{Name: entryPaste},
			{Name: entryDelete, Shortcut: "Del"},
			{Separator: true},
			{Name: entryMoveUp, Shortcut: "Shift+Up"},
			{Name: entryMoveDown, Shortcut: "Shift+Down"},
		}},
		{Name: "Run", Entries: []menubar.Entry{
			{Name: entryStart},
			{Name: entryReview},
			{Separator: true},
			{Name: entryPause},
			{Name: entryResume},
			{Name: entryStop},
			{Name: entryRetry},
			{Separator: true},
			{Name: entryAddLink},
			{Name: entryReviewSongs},
		}},
		{Name: "Help", Entries: []menubar.Entry{
			{Name: entryAbout},
		}},
	}
}
