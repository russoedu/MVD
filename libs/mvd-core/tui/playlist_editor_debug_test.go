//go:build mvddebug

package tui

import (
	"bufio"
	tea "charm.land/bubbletea/v2"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	editorpage "youtube-downloader/libs/mvd-core/editor/editor"

	"youtube-downloader/libs/mvd-core/playlisteditor"
)

// These tests run only in a build made for testing (go test -tags mvddebug), the only one
// that asks why a decision was changed.
func TestAChangedDecisionAsksWhyAndTheAnswerIsLogged(t *testing.T) {
	f := newEditorFixture(t, samePlan())
	f.planDone()

	// Skipping goes against the proposal, so it asks.
	f.press("x")
	if f.editor.Mode() != editorpage.Input || !strings.Contains(f.screen(), "Why?") {
		t.Fatalf("the reason should be asked for:\n%s", f.screen())
	}
	f.drive(tea.PasteMsg{Content: "this is a live version"})
	f.drive(key("enter"))
	if f.editor.Mode() != editorpage.Browsing {
		t.Fatal("the question should close")
	}

	// Confirming agrees with the proposal: nothing to ask.
	f.press("down", "c")
	if f.editor.Mode() != editorpage.Browsing {
		t.Error("confirming should not ask why")
	}

	file, err := os.Open(filepath.Join(f.dir, "decisions.jsonl"))
	if err != nil {
		t.Fatalf("the log was not written: %v", err)
	}
	defer func() { _ = file.Close() }()
	var recs []playlisteditor.DecisionRecord
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		var rec playlisteditor.DecisionRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatal(err)
		}
		recs = append(recs, rec)
	}
	if len(recs) != 1 || recs[0].Decision != "skipped" || recs[0].Note != "this is a live version" || recs[0].Proposed != "OFFICIAL001" || recs[0].Reason != "because" {
		t.Errorf("records %+v", recs)
	}
}
