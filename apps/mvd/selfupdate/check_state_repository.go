package selfupdate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const stateFile = "update-check.json"

// checkState is what is remembered between starts.
type checkState struct {
	// CheckedAt is when GitHub was last asked.
	CheckedAt time.Time `json:"checked_at"`
	// Latest is the newest version GitHub named then.
	Latest string `json:"latest"`
	// Told is the version the person was already told about, so it is said once.
	Told string `json:"told"`
}

func loadState(dir string) checkState {
	var s checkState
	if data, err := os.ReadFile(filepath.Join(dir, stateFile)); err == nil {
		_ = json.Unmarshal(data, &s)
	}

	return s
}

func saveState(dir string, s checkState) {
	if data, err := json.Marshal(s); err == nil {
		_ = os.WriteFile(filepath.Join(dir, stateFile), data, 0o600)
	}
}
