// Package tui shows the run on a fixed full screen interface: counters,
// playlists, entries, the output of the selected item and a key bar.
package tui

import (
	"github.com/charmbracelet/lipgloss"

	"youtube-downloader/libs/mvd-core/config"
)

// Colours and styles, rebuilt from the config by ApplyTheme. They start as the
// palette taken from the logo.
var (
	colMagenta, colCyan, colYellow, colGreen, colRed, colDim, colText lipgloss.Color

	styTitle, styBorder, styBorderOn, styDim, styText, styCyan, styYellow lipgloss.Style
	styGreen, styRed, styMagenta, stySelected, styKey, styBarFill         lipgloss.Style
	styBarEmpty                                                           lipgloss.Style
)

func init() { ApplyTheme(config.DefaultThemeColors()) }

// ApplyTheme sets the interface colours. Call it before the first screen is
// built; it is not safe to call while a program is rendering.
func ApplyTheme(c config.ThemeColors) {
	colMagenta = lipgloss.Color(c.Accent)
	colCyan = lipgloss.Color(c.Focus)
	colYellow = lipgloss.Color(c.Highlight)
	colGreen = lipgloss.Color(c.Success)
	colRed = lipgloss.Color(c.Error)
	colDim = lipgloss.Color(c.Dim)
	colText = lipgloss.Color(c.Text)

	styTitle = lipgloss.NewStyle().Bold(true).Foreground(colMagenta)
	styBorder = lipgloss.NewStyle().Foreground(colDim)
	styBorderOn = lipgloss.NewStyle().Foreground(colCyan)
	styDim = lipgloss.NewStyle().Foreground(colDim)
	styText = lipgloss.NewStyle().Foreground(colText)
	styCyan = lipgloss.NewStyle().Foreground(colCyan)
	styYellow = lipgloss.NewStyle().Foreground(colYellow)
	styGreen = lipgloss.NewStyle().Foreground(colGreen)
	styRed = lipgloss.NewStyle().Foreground(colRed)
	styMagenta = lipgloss.NewStyle().Foreground(colMagenta)
	stySelected = lipgloss.NewStyle().Background(lipgloss.Color(c.Selected)).Foreground(colText).Bold(true)
	styKey = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	styBarFill = lipgloss.NewStyle().Foreground(colYellow)
	styBarEmpty = lipgloss.NewStyle().Foreground(colDim)
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	minWideWidth = 100
	headerHeight = 3
	keyBarHeight = 1
)
