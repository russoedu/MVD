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
	{"mac-os-7", "Mac OS 7.6"},
	{"mac-os-9", "Mac OS 9"},
	{"aqua", "Mac OS X 10.0"},
	{"mac-os-x-10-4", "Mac OS X 10.4"},
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

// WindowScheme is a colour scheme of a classic window style, as the system offered it in its
// Appearance or Display settings: ID is what the config file keeps and the page's style is
// named after, Name what the preferences show.
type WindowScheme struct {
	ID, Name string
}

// WindowSchemes are the colour schemes of each classic style, the system's own first. The terminal
// look has none: its colours are the themes of the terminal look.
var WindowSchemes = map[string][]WindowScheme{
	"windows-3-1": {
		{"default", "Windows Default"}, {"hot-dog-stand", "Hot Dog Stand"}, {"ocean", "Ocean"}, {"monochrome", "Monochrome"},
	},
	"windows-95": {
		{"standard", "Windows Standard"}, {"high-contrast-black", "High Contrast Black"}, {"high-contrast-white", "High Contrast White"},
		{"rainy-day", "Rainy Day"}, {"brick", "Brick"},
	},
	"windows-xp": {
		{"blue", "Default (blue)"}, {"olive-green", "Olive Green"}, {"silver", "Silver"},
	},
	"windows-vista": {
		{"aero", "Windows Aero"}, {"high-contrast-black", "High Contrast Black"},
	},
	"mac-os-7": {
		{"standard", "Standard"}, {"black-and-white", "Black & White"}, {"teal", "Teal"}, {"rose", "Rose"},
	},
	"mac-os-9": {
		{"default", "Mac OS Default"}, {"bondi-blue", "Bondi Blue"}, {"grape", "Grape"}, {"tangerine", "Tangerine"}, {"graphite", "Graphite"},
	},
	"aqua":          {{"blue", "Blue"}, {"graphite", "Graphite"}},
	"mac-os-x-10-4": {{"blue", "Blue"}, {"graphite", "Graphite"}},
}

// WindowSchemeIDs are the ids of the colour schemes of style, in order; none for the terminal look.
func WindowSchemeIDs(style string) []string {
	schemes := WindowSchemes[style]
	ids := make([]string, len(schemes))
	for i, s := range schemes {
		ids[i] = s.ID
	}

	return ids
}

// WindowSchemeName is how the preferences name a scheme of style; an unknown id reads as itself.
func WindowSchemeName(style, id string) string {
	for _, s := range WindowSchemes[style] {
		if s.ID == id {
			return s.Name
		}
	}

	return id
}

// WindowSchemeOf is the scheme of style the settings mean: scheme when style has it, else the
// style's first, or "" for a style with none.
func WindowSchemeOf(style, scheme string) string {
	schemes := WindowSchemes[style]
	for _, s := range schemes {
		if s.ID == scheme {
			return scheme
		}
	}
	if len(schemes) == 0 {
		return ""
	}

	return schemes[0].ID
}
