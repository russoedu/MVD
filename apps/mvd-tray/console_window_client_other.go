//go:build !windows

package main

// attachParentConsole does nothing here: only a windowed Windows program lacks the
// terminal that started it.
func attachParentConsole() {}

// showFatal does nothing here: the error was already printed to the terminal.
func showFatal(string) {}

// askYesNo says no here without asking: nothing on these systems offers the question.
func askYesNo(string, string) bool { return false }

// trayUnavailable does nothing here: the terminal says what happened and how to quit.
func trayUnavailable(string) {}
