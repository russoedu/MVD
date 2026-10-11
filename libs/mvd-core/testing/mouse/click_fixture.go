// Package mouse makes the mouse events the tests of the screens send.
package mouse

import tea "charm.land/bubbletea/v2"

// Click is a press of the left button at column x and row y.
func Click(x, y int) tea.MouseMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

// Wheel is a notch of the wheel at column x and row y.
func Wheel(x, y int, button tea.MouseButton) tea.MouseMsg {
	return tea.MouseWheelMsg{X: x, Y: y, Button: button}
}
