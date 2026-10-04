package appdir

import (
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(d) != "mvd" {
		t.Errorf("app dir should end in mvd, got %q", d)
	}
}

func TestDefaultDownloadsDir(t *testing.T) {
	// Just assert it returns a non-empty absolute-ish path without panicking.
	if DefaultDownloadsDir() == "" {
		t.Error("downloads dir should not be empty")
	}
}
