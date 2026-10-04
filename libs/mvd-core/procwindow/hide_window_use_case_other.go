//go:build !windows

package procwindow

import "os/exec"

// Hide does nothing here: only Windows opens a console window for a child process
// when its parent has none.
func Hide(*exec.Cmd) {}
