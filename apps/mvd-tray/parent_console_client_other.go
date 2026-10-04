//go:build !windows

package main

// attachParentConsole does nothing here: only a windowed Windows program lacks the
// terminal that started it.
func attachParentConsole() {}
