//go:build !windows

package main

// hideOwnConsole does nothing here: only Windows opens a console window next to a
// program that lives in the tray.
func hideOwnConsole() bool {
	return false
}
