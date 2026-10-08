package config

import (
	"regexp"
	"strings"
)

// ThemeColors are the colours of the terminal interface, as #rgb or #rrggbb.
type ThemeColors struct {
	// Accent colours the title and the magenta highlights.
	Accent string
	// Focus colours the border of the focused panel and the cyan details.
	Focus string
	// Highlight colours the key hints and the progress bar.
	Highlight string
	Success   string
	Error     string
	// Dim colours borders and secondary text.
	Dim  string
	Text string
	// Selected is the background of the selected row.
	Selected string
}

// DefaultThemeColors is the palette taken from the logo.
func DefaultThemeColors() ThemeColors {
	return ThemeColors{
		Accent:    "#ff007f",
		Focus:     "#00f0ff",
		Highlight: "#ffe600",
		Success:   "#3ddc84",
		Error:     "#ff4d4d",
		Dim:       "#6b7280",
		Text:      "#e5e7eb",
		Selected:  "#3b0f2a",
	}
}

var hexColour = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// validColour reports whether val is a #rgb or #rrggbb colour.
func validColour(val string) bool {
	return hexColour.MatchString(strings.TrimSpace(val))
}
