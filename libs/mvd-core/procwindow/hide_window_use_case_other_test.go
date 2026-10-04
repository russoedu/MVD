//go:build !windows

package procwindow

import (
	"os/exec"
	"testing"
)

func TestHideChangesNothingOutsideWindows(t *testing.T) {
	cmd := exec.Command("anything")

	Hide(cmd)

	if cmd.SysProcAttr != nil {
		t.Errorf("SysProcAttr was set: %+v", cmd.SysProcAttr)
	}
}
