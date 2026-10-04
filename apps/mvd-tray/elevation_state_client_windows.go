package main

import "golang.org/x/sys/windows"

// isElevated reports whether the app is running with administrator rights, which is when
// it can write to Program Files. It only asks; it never tries to obtain them.
func isElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
