//go:build !windows

package install

import "errors"

// createShortcut is not available here: only Windows has the Start menu shortcuts.
func createShortcut(_, _ string) error {
	return errors.New("shortcuts are made on Windows only")
}
