package terminalui

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	ttygo "github.com/meta-tui/treactui/packages/tty-go"
)

// ConfigPath is the settings file, the same one the terminal app uses.
func ConfigPath(appDir string) string {
	return filepath.Join(appDir, "config.conf")
}

// NewShared prepares the app as one program that the window shows (and that a reload of
// the page finds as it was). newModel is called when the page first connects, and again
// for the first connection after the user quits the app from it.
func NewShared(newModel func() tea.Model) *ttygo.SharedProgram {
	return ttygo.NewSharedProgram(newModel)
}
