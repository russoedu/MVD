//go:build !windows

package main

// trayUnavailable does nothing here: the terminal says what happened and how to quit.
func trayUnavailable(string) {}
