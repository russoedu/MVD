package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestARenamedProgramLeavesItsFolderFreeToBeRemoved(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "MVD")
	program := filepath.Join(folder, "mvd-tray.exe")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(program, []byte("program"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(filepath.Join(os.TempDir(), fmt.Sprintf("mvd-removed-%d.exe", os.Getpid()))) })

	leftover, err := removeRunningProgram(program)

	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(program); err == nil {
		t.Error("the program is still in its folder")
	}
	if leftover != "" {
		t.Errorf("leftover = %s, want none when the temporary folder could take it", leftover)
	}
	if err := os.RemoveAll(folder); err != nil {
		t.Errorf("the folder could not be removed: %v", err)
	}
}

func TestAProgramThatIsNotThereIsAnError(t *testing.T) {
	if _, err := removeRunningProgram(filepath.Join(t.TempDir(), "missing.exe")); err == nil {
		t.Error("expected an error")
	}
}
