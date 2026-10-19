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
			{Name: "Undo", Shortcut: "Ctrl+Z"},
			{Name: "Redo", Shortcut: "Ctrl+Y"},
			{Separator: true},
			{Name: "Copy", Shortcut: "Ctrl+C"},
			{Name: "Paste", Shortcut: "Ctrl+V"},
			{Name: "Delete", Shortcut: "Del"},
			{Separator: true},
			{Name: "Move up"},
			{Name: "Move down"},
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
	"Undo", "Redo", "Copy", "Paste", "Delete", "Move up", "Move down",
}
