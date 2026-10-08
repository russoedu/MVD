//go:build !windows

package nativewindow

// Open is not available here yet: only the Windows web view is built so far, so
// the caller falls back to a browser.
func Open(_, _ string) error { return ErrUnavailable }
