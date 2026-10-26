package config

import (
	"path/filepath"
	"testing"
)

func TestTheWindowStyleIsKeptInTheFileAndAnUnknownOneIsIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.conf")
	cfg := Default(dir, dir)
	if cfg.WindowStyle != WindowStyleDOS {
		t.Fatalf("a new config starts with %q", cfg.WindowStyle)
	}
	cfg.WindowStyle = "mac-os-9"
	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}
	back, _, err := LoadOrCreate(path, dir, dir)
	if err != nil || back.WindowStyle != "mac-os-9" {
		t.Fatalf("read back %q, %v", back.WindowStyle, err)
	}
	if ValidWindowStyle("amiga") || !ValidWindowStyle("windows-3-1") {
		t.Error("the styles are not the ones offered")
	}
	if WindowStyleName("aqua") != "Mac OS X 10.0" {
		t.Errorf("aqua reads %q", WindowStyleName("aqua"))
	}
}
