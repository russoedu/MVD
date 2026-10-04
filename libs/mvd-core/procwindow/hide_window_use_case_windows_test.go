package procwindow

import (
	"os/exec"
	"testing"
)

// noWindowFlag is CREATE_NO_WINDOW.
const noWindowFlag = 0x08000000

func TestHideStartsAChildWithoutAConsoleWindow(t *testing.T) {
	cmd := exec.Command("anything")

	Hide(cmd)

	if cmd.SysProcAttr == nil || cmd.SysProcAttr.CreationFlags&noWindowFlag == 0 {
		t.Fatalf("CREATE_NO_WINDOW is not set: %+v", cmd.SysProcAttr)
	}
	if cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow must stay off, or a graphical child such as the folder chooser would be hidden too")
	}
}

func TestHideKeepsFlagsThatWereAlreadySetAndIsSafeToRepeat(t *testing.T) {
	cmd := exec.Command("anything")
	Hide(cmd)
	cmd.SysProcAttr.CreationFlags |= 0x00000200 // CREATE_NEW_PROCESS_GROUP

	Hide(cmd)

	if want := uint32(noWindowFlag | 0x00000200); cmd.SysProcAttr.CreationFlags != want {
		t.Errorf("flags = %#x, want %#x", cmd.SysProcAttr.CreationFlags, want)
	}
}

func TestAHiddenChildReallyRunsAndReturnsItsOutput(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "echo", "hello")
	Hide(cmd)

	out, err := cmd.Output()

	if err != nil || len(out) == 0 {
		t.Fatalf("output %q, error %v", out, err)
	}
}
