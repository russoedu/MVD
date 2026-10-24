//go:build windows

package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAShortcutIsMadeWithoutStartingAnyProgram(t *testing.T) {
	link := filepath.Join(t.TempDir(), `Menu; calc #'$(a) %PATH%`, "MVD.lnk")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := createShortcut(link, `C:\Program Files\MVD\mvd.exe`); err != nil {
		t.Fatal(err)
	}

	if info, err := os.Stat(link); err != nil || info.Size() == 0 {
		t.Errorf("no shortcut file: %v", err)
	}
}
