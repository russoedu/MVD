package procwindow

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// Hide makes cmd start without a console window of its own.
//
// Only CREATE_NO_WINDOW is set, not HideWindow: HideWindow would also hide the first
// window a graphical child opens, such as the folder chooser.
func Hide(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= windows.CREATE_NO_WINDOW
}
