package deps

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
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

	dir := t.TempDir()
	archive := filepath.Join(dir, "ffmpeg.zip")
	if err := os.WriteFile(archive, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "ffmpeg")
	if err := extractZipFile(archive, "ffmpeg", dest); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(dest); string(data) != "binary" {
		t.Errorf("unexpected content %q", data)
	}
	if err := extractZipFile(archive, "nope", dest); err == nil {
		t.Error("missing entry should fail")
	}
	if err := extractZipFile(filepath.Join(dir, "missing.zip"), "ffmpeg", dest); err == nil {
		t.Error("a missing archive should fail")
	}
}

func TestYtDlpIsAlwaysKeptUpToDateFromItsOwnStandaloneBuilds(t *testing.T) {
	want := map[[2]string]string{
		{"windows", "amd64"}: "yt-dlp.exe",
		{"windows", "arm64"}: "yt-dlp_arm64.exe",
		{"darwin", "arm64"}:  "yt-dlp_macos",
		{"darwin", "amd64"}:  "yt-dlp_macos",
		{"linux", "amd64"}:   "yt-dlp_linux",
		{"linux", "arm64"}:   "yt-dlp_linux_aarch64",
	}
	for platform, asset := range want {
		dep := platformDeps(platform[0], platform[1])["yt-dlp"]

		if dep.Source == nil || dep.Source.Repo != "yt-dlp/yt-dlp" || dep.Source.Asset != asset {
			t.Errorf("%v: source = %+v, want yt-dlp/yt-dlp %s", platform, dep.Source, asset)
		}
		if dep.MinAge != 0 {
			t.Errorf("%v: yt-dlp must be replaced by any newer build, MinAge = %v", platform, dep.MinAge)
		}
		if dep.IsZip {
			t.Errorf("%v: the yt-dlp builds are single programs", platform)
		}
	}
	if got := platformDeps("linux", "amd64")["yt-dlp"].URL; got != "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux" {
		t.Errorf("the Linux build must be the standalone one, which needs no Python; got %s", got)
	}
}

func TestFfmpegIsKeptUpToDateOnWindowsOnlyAndNotMoreOftenThanEveryThirtyDays(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		dep := platformDeps("windows", arch)["ffmpeg"]

		if dep.Source == nil || dep.Source.Repo != "yt-dlp/FFmpeg-Builds" || !dep.IsZip {
			t.Errorf("windows/%s: %+v", arch, dep)
		}
		if dep.MinAge != 30*24*time.Hour {
			t.Errorf("windows/%s: MinAge = %v, want 30 days, because the builds are republished daily", arch, dep.MinAge)
		}
	}
	for _, goos := range []string{"darwin", "linux"} {
		if dep := platformDeps(goos, "amd64")["ffmpeg"]; dep.Source != nil {
			t.Errorf("%s has no published build we can unpack, so ffmpeg is fetched once: %+v", goos, dep)
		}
	}
}

func TestDenoIsFetchedOnceAndNeverUpdated(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		if dep := platformDeps(goos, "amd64")["deno"]; dep.Source != nil {
			t.Errorf("%s: deno should not be managed: %+v", goos, dep)
		}
	}
}
