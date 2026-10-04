//go:build !windows

package install

// isElevated is false here: only Windows has a place for everyone that needs it.
func isElevated() bool { return false }
