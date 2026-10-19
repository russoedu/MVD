//go:build !windows && !linux

package filetype

// Register does nothing here: macOS learns about the kind of file from the Info.plist of the
// bundle.
func Register(string) error { return nil }

// Unregister does nothing here, for the same reason.
func Unregister() error { return nil }
