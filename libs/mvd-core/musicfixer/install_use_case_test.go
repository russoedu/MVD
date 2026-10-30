package musicfixer

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	_ = zw.Close()
	return buf.Bytes()
}

func exeName() string {
	if runtime.GOOS == "windows" {
		return "music-fixer/music-fixer.exe"
	}
	return "music-fixer/music-fixer"
}

func TestUnpackPutsTheProgramInPlaceAndKeepsItForTheSameBundle(t *testing.T) {
	dir := t.TempDir()
	data := makeZip(t, map[string]string{exeName(): "v1", "music-fixer/app.json": "{}"})

	exe, err := unpack(data, dir)
	if err != nil || filepath.Dir(exe) != filepath.Join(dir, "music-fixer") {
		t.Fatalf("exe %q, %v", exe, err)
	}
	_ = os.WriteFile(exe, []byte("edited"), 0o755)
	if _, err := unpack(data, dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "edited" {
		t.Errorf("the same bundle was unpacked again: %q", got)
	}

	if _, err := unpack(makeZip(t, map[string]string{exeName(): "v2"}), dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "v2" {
		t.Errorf("a new bundle was not unpacked: %q", got)
	}
}

func TestUnpackRefusesAFileOutsideItsFolder(t *testing.T) {
	if _, err := unpack(makeZip(t, map[string]string{"../evil": "x"}), t.TempDir()); err == nil {
		t.Error("a path going up was accepted")
	}
}

func TestInstallWithoutABundleSaysSo(t *testing.T) {
	if _, err := Install(t.TempDir()); err != ErrNotBundled {
		t.Errorf("err: %v", err)
	}
}
