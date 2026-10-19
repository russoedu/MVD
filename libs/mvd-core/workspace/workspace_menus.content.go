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
	entryEditList    = "Edit list…"
	entryUndo        = "Undo"
	entryRedo        = "Redo"
	entryDelete      = "Delete"
	entryMoveUp      = "Move up"
	entryMoveDown    = "Move down"
	entryAbout       = "About MVD"
	entryOpenSource  = "Open Source"
)

// openSourceURL is where the code of the app is.
const openSourceURL = "https://github.com/russoedu/mvd"

func menus() []menubar.Menu {
	return []menubar.Menu{
		{Name: "File", Entries: []menubar.Entry{
			{Name: entryNew, Shortcut: "Ctrl+N"},
			{Name: entryOpen, Shortcut: "Ctrl+O"},
			{Name: entrySave, Shortcut: "Ctrl+S"},
			{Name: entrySaveAs},
			{Separator: true},
			{Name: entryPreferences, Shortcut: "Ctrl+,"},
			{Separator: true},
			{Name: entryQuit, Shortcut: "Ctrl+Q"},
		}},
		{Name: "Edit", Entries: []menubar.Entry{
			{Name: entryEditList, Shortcut: "Ctrl+E"},
			{Separator: true},
			{Name: entryUndo, Shortcut: "Ctrl+Z"},
			{Name: entryRedo, Shortcut: "Ctrl+Y"},
			{Separator: true},
			{Name: "Copy", Shortcut: "Ctrl+C"},
			{Name: "Paste", Shortcut: "Ctrl+V"},
			{Name: entryDelete, Shortcut: "Del"},
			{Separator: true},
			{Name: entryMoveUp, Shortcut: "Shift+Up"},
			{Name: entryMoveDown, Shortcut: "Shift+Down"},
		}},
		{Name: "Help", Entries: []menubar.Entry{
			{Name: entryAbout},
			{Name: entryOpenSource},
		}},
	}
}

// notYet are the entries whose popups come with later parts of the list rework; they are
// shown greyed out until then.
var notYet = []string{
	"Copy", "Paste",
}
