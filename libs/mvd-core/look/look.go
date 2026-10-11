// Package look holds the interface colours and the styles made from them. They start as
// the palette taken from the logo and are rebuilt from the config by Apply.
package look

import (
	"charm.land/lipgloss/v2"

	"youtube-downloader/libs/mvd-core/config"
)

// Colours are the interface colours Apply last set; the screens made from Cuppa designs
// take them as their palette.
var Colours = config.DefaultThemeColors()

// The styles of the interface, rebuilt by Apply.
var (
	Title, Border, BorderOn, Dim, Text, Cyan, Yellow lipgloss.Style
	Green, Red, Key                                  lipgloss.Style
)

func init() { Apply(config.DefaultThemeColors()) }

// Apply sets the interface colours. Call it before the first screen is built; it is not
// safe to call while a program is rendering.
func Apply(c config.ThemeColors) {
	Colours = c
	accent := lipgloss.Color(c.Accent)
	focus := lipgloss.Color(c.Focus)
	highlight := lipgloss.Color(c.Highlight)
	dim := lipgloss.Color(c.Dim)

	Title = lipgloss.NewStyle().Bold(true).Foreground(accent)
	Border = lipgloss.NewStyle().Foreground(dim)
	BorderOn = lipgloss.NewStyle().Foreground(focus)
	Dim = lipgloss.NewStyle().Foreground(dim)
	Text = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Text))
	Cyan = lipgloss.NewStyle().Foreground(focus)
	Yellow = lipgloss.NewStyle().Foreground(highlight)
	Green = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Success))
	Red = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Error))
	Key = lipgloss.NewStyle().Foreground(highlight).Bold(true)
}

// Palette has the fields of the Palette every screen of the designs has, so a screen's own
// Palette type is a conversion away.
type Palette struct {
	Accent, Focus, Highlight, Success, Error, Dim, Text, Selected string
}

// CurrentPalette returns the current interface colours as the palette of the screens.
func CurrentPalette() Palette {
	c := Colours
	return Palette{
		Accent: c.Accent, Focus: c.Focus, Highlight: c.Highlight, Success: c.Success,
		Error: c.Error, Dim: c.Dim, Text: c.Text, Selected: c.Selected,
	}
}
