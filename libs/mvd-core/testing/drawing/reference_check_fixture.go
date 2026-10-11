package drawing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Check compares got, a screen w cells wide, with the reference drawing dir/name.ansi.
// UPDATE_LEGACY=name,name rewrites those references from got: only for a screen that
// was never migrated, or the reference is lost.
func Check(t *testing.T, dir, name string, w int, got string) {
	t.Helper()
	path := filepath.Join(dir, name+".ansi")
	if wanted := os.Getenv("UPDATE_LEGACY"); wanted != "" && strings.Contains(","+wanted+",", ","+name+",") {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s: no reference drawing (%v); run with UPDATE_LEGACY=%s", name, err, name)
		return
	}
	if diffs := SameCells(string(want), got, w, 8); len(diffs) > 0 {
		t.Errorf("%s differs from its reference:\n  %s", name, strings.Join(diffs, "\n  "))
	}
}
