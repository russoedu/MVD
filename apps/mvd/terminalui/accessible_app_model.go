package terminalui

import (
	tea "charm.land/bubbletea/v2"

	ttygo "github.com/meta-ui-labs/treactui/packages/tty-go"

	"youtube-downloader/libs/mvd-core/tui"
)

type outliner interface{ Outline() tui.ScreenOutline }

// accessibleApp adds ttygo.Accessible to mvd's own app model without changing
// it: Init and View are mvd's, Update keeps the wrapper in place.
type accessibleApp struct{ tea.Model }

// Update passes msg to the wrapped model and returns the next model, wrapped again, with its command.
func (a accessibleApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := a.Model.Update(msg)
	return accessibleApp{next}, cmd
}

// Accessible returns the snapshot of the current screen, built from the wrapped model's outline. It panics when the wrapped model has no Outline method.
func (a accessibleApp) Accessible() ttygo.Snapshot {
	return snapshotFromOutline(a.Model.(outliner).Outline())
}
