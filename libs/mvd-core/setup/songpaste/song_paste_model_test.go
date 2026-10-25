package songpaste

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTheSummaryCountsSongsAndLinesNotUnderstood(t *testing.T) {
	p := New(t.TempDir())
	if got := p.Summary(); got != "0 songs" {
		t.Errorf("an empty paste says %q", got)
	}
	p.SetText("Daft Punk - One More Time\nbad line\nAlexia - Uh La La La")
	if got := p.Summary(); got != "2 songs, 1 line not understood" {
		t.Errorf("two songs and a bad line say %q", got)
	}
}

func TestSavingWritesTheFileAndNamesIt(t *testing.T) {
	dir := t.TempDir()
	p := New(dir)
	p.SetText("Daft Punk - One More Time")
	if out := p.Save(time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)); out != Added {
		t.Fatalf("saving gave %v, error %q", out, p.Err)
	}
	if filepath.Dir(p.Saved) != dir {
		t.Errorf("saved in %q, want inside %q", p.Saved, dir)
	}
	if _, err := os.Stat(p.Saved); err != nil {
		t.Errorf("the saved file is not there: %v", err)
	}
}

func TestAFailedSaveShowsWhyUntilTheNextThingTyped(t *testing.T) {
	// A file where the folder should be, so the folder cannot be made.
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	p := New(filepath.Join(blocker, "lists"))
	p.SetText("Daft Punk - One More Time")
	if out := p.Save(time.Now()); out != None || p.Err == "" {
		t.Fatalf("a save that cannot be written gave %v with error %q", out, p.Err)
	}
	if p.Summary() != p.Err {
		t.Error("the error should replace the summary")
	}
	p.ClearError()
	if p.Err != "" {
		t.Error("typing should forget the error")
	}
}

func TestLeavingDoesNotSave(t *testing.T) {
	p := New(t.TempDir())
	if p.Leave() != Cancel || p.Saved != "" {
		t.Error("leaving should cancel and save nothing")
	}
}
