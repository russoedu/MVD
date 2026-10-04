// Package tui shows the run on a fixed full screen interface: counters,
// playlists, entries, the output of the selected item and a key bar.
package tui

import "github.com/charmbracelet/lipgloss"

// Palette taken from the logo.
var (
	colMagenta = lipgloss.Color("#ff007f")
	colCyan    = lipgloss.Color("#00f0ff")
	colYellow  = lipgloss.Color("#ffe600")
	colGreen   = lipgloss.Color("#3ddc84")
	colRed     = lipgloss.Color("#ff4d4d")
	colDim     = lipgloss.Color("#6b7280")
	colText    = lipgloss.Color("#e5e7eb")

	styTitle    = lipgloss.NewStyle().Bold(true).Foreground(colMagenta)
	styBorder   = lipgloss.NewStyle().Foreground(colDim)
	styBorderOn = lipgloss.NewStyle().Foreground(colCyan)
	styDim      = lipgloss.NewStyle().Foreground(colDim)
	styText     = lipgloss.NewStyle().Foreground(colText)
	styCyan     = lipgloss.NewStyle().Foreground(colCyan)
	styYellow   = lipgloss.NewStyle().Foreground(colYellow)
	styGreen    = lipgloss.NewStyle().Foreground(colGreen)
	styRed      = lipgloss.NewStyle().Foreground(colRed)
	styMagenta  = lipgloss.NewStyle().Foreground(colMagenta)
	stySelected = lipgloss.NewStyle().Background(lipgloss.Color("#3b0f2a")).Foreground(colText).Bold(true)
	styKey      = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	styBarFill  = lipgloss.NewStyle().Foreground(colYellow)
	styBarEmpty = lipgloss.NewStyle().Foreground(colDim)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	minWideWidth = 100
	headerHeight = 3
	keyBarHeight = 1
)
