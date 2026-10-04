//go:build !windows

package uninstall

import "os"

// removeRunningProgram deletes the running program: on these systems a program that is
// running can be unlinked, and it keeps running from memory until it exits.
func removeRunningProgram(path string) (leftover string, err error) {
	return "", os.Remove(path)
}
