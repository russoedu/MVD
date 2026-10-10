package playlisteditor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

// DecisionRecord is one line of the decision log: what the algorithm proposed for a song,
// what the person chose instead, and why they said they did.
type DecisionRecord struct {
	Time     time.Time `json:"time"`
	Playlist string    `json:"playlist"`
	Song     string    `json:"song"`
	UploadID string    `json:"upload_id"`
	// Proposed is what the algorithm chose, and ProposedKind and Reason say how.
	Proposed     string `json:"proposed"`
	ProposedKind string `json:"proposed_kind"`
	Reason       string `json:"reason,omitempty"`
	// Decision is what the person did (replaced, original, skipped), Chosen the video they
	// took, if any, and Note what they wrote.
	Decision string `json:"decision"`
	Chosen   string `json:"chosen,omitempty"`
	Note     string `json:"note,omitempty"`
}

// RecordFor makes the log line for a decision on a row, which already holds the decision.
func RecordFor(e playlistfile.Entry, note string) DecisionRecord {
	return DecisionRecord{
		Time:         time.Now().UTC(),
		Playlist:     e.Playlist,
		Song:         e.Upload.Title,
		UploadID:     e.Upload.ID,
		Proposed:     e.TargetID,
		ProposedKind: e.Kind,
		Reason:       e.Reason,
		Decision:     e.Decision,
		Chosen:       e.ChosenID,
		Note:         note,
	}
}

// AppendDecision adds a record to the log at path, one JSON line each. The log stays on
// this machine: nothing reads it but the person who is improving the algorithm.
func AppendDecision(path string, rec DecisionRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	line, err := json.Marshal(rec)
	if err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
