package terminalui

import "github.com/atotto/clipboard"

// SystemClipboard is the machine's own clipboard, which the window shows its terminal in
// cannot reach.
type SystemClipboard struct{}

// Read is the text on the clipboard.
func (SystemClipboard) Read() (string, error) { return clipboard.ReadAll() }

// Write puts text on the clipboard.
func (SystemClipboard) Write(text string) error { return clipboard.WriteAll(text) }
