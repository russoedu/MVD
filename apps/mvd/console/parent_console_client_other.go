//go:build !windows

package console

// AttachParent does nothing here: only a windowed Windows program lacks the
// terminal that started it.
func AttachParent() {}
