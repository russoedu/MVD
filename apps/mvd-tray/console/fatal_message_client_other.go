//go:build !windows

package console

// ShowFatal does nothing here: the error was already printed to the terminal.
func ShowFatal(string) {}
