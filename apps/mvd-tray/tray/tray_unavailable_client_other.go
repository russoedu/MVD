//go:build !windows

package tray

// Unavailable does nothing here: the terminal says what happened and how to quit.
func Unavailable(string) {}
