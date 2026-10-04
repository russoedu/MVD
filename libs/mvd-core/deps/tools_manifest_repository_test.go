package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTheManifestSurvivesBeingSavedAndLoaded(t *testing.T) {
	dir := t.TempDir()
	want := manifest{"yt-dlp": {AssetID: 7, Tag: "2026.08.19", UpdatedAt: epoch, SHA256: "abc"}}

	if err := saveManifest(dir, want); err != nil {
		t.Fatal(err)
	}
	got := loadManifest(dir)

	if got["yt-dlp"] != want["yt-dlp"] {
		t.Errorf("got %+v, want %+v", got["yt-dlp"], want["yt-dlp"])
	}
	if listing := listing(t, dir); listing != manifestName {
		t.Errorf("folder holds %q, want only the manifest (no temporary file)", listing)
	}
}

func TestAManifestWithAByteOrderMarkIsStillRead(t *testing.T) {
	dir := t.TempDir()
	json := "\xef\xbb\xbf" + `{"yt-dlp": {"assetId": 5, "tag": "2026.08.19"}}`
	if err := os.WriteFile(filepath.Join(dir, manifestName), []byte(json), 0o600); err != nil {
		t.Fatal(err)
	}

	got := loadManifest(dir)

	if got["yt-dlp"].AssetID != 5 || got["yt-dlp"].Tag != "2026.08.19" {
		t.Errorf("a file saved by PowerShell or an editor was not read: %+v", got)
	}
}

func TestAMissingOrDamagedManifestIsAnEmptyOneNotACrash(t *testing.T) {
	dir := t.TempDir()
	if got := loadManifest(dir); len(got) != 0 {
		t.Errorf("missing manifest = %v", got)
	}
	for _, damaged := range []string{"{not json", "null", ""} {
		if err := os.WriteFile(filepath.Join(dir, manifestName), []byte(damaged), 0o600); err != nil {
			t.Fatal(err)
		}
		got := loadManifest(dir)
		if got == nil {
			t.Errorf("%q gave a nil manifest that would panic on write", damaged)
		}
		if len(got) != 0 {
			t.Errorf("%q gave %v", damaged, got)
		}
	}
}
