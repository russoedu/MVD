package terminalui

import (
	tea "charm.land/bubbletea/v2"

	ttygo "github.com/meta-tui/treactui/packages/tty-go"

	"youtube-downloader/libs/mvd-core/tui"
)

type outliner interface{ Outline() tui.ScreenOutline }

// accessibleApp adds ttygo.Accessible to mvd's own app model without changing
// it: Init and View are mvd's, Update keeps the wrapper in place.
type accessibleApp struct{ tea.Model }

func (a accessibleApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := a.Model.Update(msg)
	return accessibleApp{next}, cmd
}

func (a accessibleApp) Accessible() ttygo.Snapshot {
	return snapshotFromOutline(a.Model.(outliner).Outline())
}
