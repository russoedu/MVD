//go:build !windows

package main

// attachParentConsole does nothing here: only a windowed Windows program lacks the
// terminal that started it.
func attachParentConsole() {}

// showFatal does nothing here: the error was already printed to the terminal.
func showFatal(string) {}

// trayUnavailable does nothing here: the terminal says what happened and how to quit.
func trayUnavailable(string) {}
