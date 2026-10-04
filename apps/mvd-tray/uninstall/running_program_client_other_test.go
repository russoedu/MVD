//go:build !windows

package uninstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTheProgramIsDeletedAndNothingIsLeftBehind(t *testing.T) {
	program := filepath.Join(t.TempDir(), "mvd-tray")
	if err := os.WriteFile(program, []byte("program"), 0o755); err != nil {
		t.Fatal(err)
	}

	leftover, err := removeRunningProgram(program)

	if err != nil || leftover != "" {
		t.Fatalf("leftover = %q, err = %v", leftover, err)
	}
	if _, err := os.Stat(program); err == nil {
		t.Error("the program is still there")
	}
}
