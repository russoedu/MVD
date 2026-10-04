package deps

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// manifestName is the file, next to the tools, that records which build of each is
// installed. A tool without an entry here is not one the app installed, and is not
// treated as its own.
const manifestName = "tools.json"

// installedTool is what is known about one tool the app installed.
type installedTool struct {
	// AssetID is the build that was installed; 0 when it was fetched without asking
	// GitHub first, which makes the next check replace it.
	AssetID   int64     `json:"assetId"`
	Tag       string    `json:"tag"`
	UpdatedAt time.Time `json:"updatedAt"`
	SHA256    string    `json:"sha256"`
}

type manifest map[string]installedTool

// loadManifest reads the manifest in dir. A missing or unreadable one is an empty one,
// which only means the tools are fetched again.
func loadManifest(dir string) manifest {
	data, err := os.ReadFile(filepath.Join(dir, manifestName))
	if err != nil {
		return manifest{}
	}
	// Windows editors, and PowerShell, put a byte-order mark at the start of a UTF-8 file.
	// JSON does not allow one, and rejecting it would make the app forget it owns its
	// tools and download them all again.
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	var m manifest
	if err := json.Unmarshal(data, &m); err != nil || m == nil {
		return manifest{}
	}

	return m
}

// saveManifest writes the manifest under a temporary name and renames it, so it is
// never seen half written.
func saveManifest(dir string, m manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	final := filepath.Join(dir, manifestName)
	partial := final + ".part"
	if err := os.WriteFile(partial, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(partial, final); err != nil {
		_ = os.Remove(partial)

		return err
	}

	return nil
}
