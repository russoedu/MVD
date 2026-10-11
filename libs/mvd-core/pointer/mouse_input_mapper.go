// Package pointer reads what the mouse did: a left click, or a notch of the wheel.
package pointer

import tea "charm.land/bubbletea/v2"

// WheelStep is how many rows one notch of the mouse wheel moves.
const WheelStep = 3

// IsLeftClick is whether msg is a press of the left button.
func IsLeftClick(msg tea.MouseMsg) bool {
	click, ok := msg.(tea.MouseClickMsg)
	return ok && click.Button == tea.MouseLeft
}

// WheelDirection is -1 for a notch up, 1 for a notch down, 0 for anything else.
func WheelDirection(msg tea.MouseMsg) int {
	wheel, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return 0
	}
	switch wheel.Button {
	case tea.MouseWheelUp:
		return -1
	case tea.MouseWheelDown:
		return 1
	}
	return 0
}
