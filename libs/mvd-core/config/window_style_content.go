package config

// WindowStyleDOS is the terminal look MVD has always had, and the style it starts with.
const WindowStyleDOS = "dos"

// WindowStyle is a way the app's window can be drawn: ID is what the config file keeps and the
// page's classic themes are called, Name what the preferences show.
type WindowStyle struct {
	ID, Name string
}

// WindowStyles are the looks of the app's window: the terminal one, then the classic systems,
// Windows first and the Mac after, each oldest first.
var WindowStyles = []WindowStyle{
	{WindowStyleDOS, "MVD terminal (DOS)"},
	{"windows-3-1", "Windows 3.1"},
	{"windows-95", "Windows 95"},
	{"windows-xp", "Windows XP"},
	{"windows-vista", "Windows Vista"},
	{"mac-os-7", "Mac OS 7"},
	{"mac-os-9", "Mac OS 9"},
	{"aqua", "Mac OS X 10.0"},
}

// WindowStyleIDs are the ids of WindowStyles, in order.
func WindowStyleIDs() []string {
	ids := make([]string, len(WindowStyles))
	for i, s := range WindowStyles {
		ids[i] = s.ID
	}

	return ids
}

// WindowStyleName is how the preferences name a style; an unknown id reads as itself.
func WindowStyleName(id string) string {
	for _, s := range WindowStyles {
		if s.ID == id {
			return s.Name
		}
	}

	return id
}

// ValidWindowStyle says whether id is one of WindowStyles.
func ValidWindowStyle(id string) bool {
	for _, s := range WindowStyles {
		if s.ID == id {
			return true
		}
	}

	return false
}
