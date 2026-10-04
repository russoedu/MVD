package deps

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPlatformDeps(t *testing.T) {
	for _, c := range [][2]string{{"windows", "amd64"}, {"darwin", "amd64"}, {"darwin", "arm64"}, {"linux", "amd64"}, {"linux", "arm64"}} {
		deps := platformDeps(c[0], c[1])
		for _, name := range []string{"yt-dlp", "ffmpeg", "deno"} {
			d, ok := deps[name]
			if !ok || d.URL == "" || d.FileName == "" {
				t.Errorf("%s/%s: missing %s", c[0], c[1], name)
			}
			if c[0] == "windows" && filepath.Ext(d.FileName) != ".exe" {
				t.Errorf("windows binary %s should end in .exe", d.FileName)
			}
		}
	}
}

func TestExtractZipFile(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("ffmpeg-4.4.1/bin/ffmpeg")
	if _, err := f.Write([]byte("binary")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "ffmpeg")
	if err := extractZipFile(buf.Bytes(), "ffmpeg", dest); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(dest); string(data) != "binary" {
		t.Errorf("unexpected content %q", data)
	}
	if err := extractZipFile(buf.Bytes(), "nope", dest); err == nil {
		t.Error("missing entry should fail")
	}
}
