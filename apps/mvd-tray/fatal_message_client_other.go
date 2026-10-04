//go:build !windows

package main

// showFatal does nothing here: the error was already printed to the terminal.
func showFatal(string) {}
