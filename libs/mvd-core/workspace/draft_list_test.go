package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/engine"
)

func newListHost(t *testing.T, fake *fakeRun) (Host, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "lists")

	return Host{ListDir: dir, Start: func(config.Config, []string, []engine.PlannedEntry, bool) (ListRun, error) { return fake, nil }}, dir
}

func TestStartingANewListKeepsADraftAndStartsAtOnce(t *testing.T) {
	host, dir := newListHost(t, newFakeRun())
	m := sized(New(host, config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m, cmd := choose(m, entryStart)
	if cmd == nil || !m.life.Starting {
		t.Fatalf("the run did not begin to start: asking %v, starting %v", m.asking, m.life.Starting)
	}
	if m.asking != askNone {
		t.Errorf("a question was asked before starting: %v", m.asking)
	}
	if _, err := os.Stat(filepath.Join(dir, "Untitled list.mvd")); err != nil {
		t.Errorf("no draft was written: %v", err)
	}
	if !strings.Contains(m.done, "draft") {
		t.Errorf("the draft was not mentioned: %q", m.done)
	}
}

func TestASecondNewListGetsAnotherDraftName(t *testing.T) {
	host, dir := newListHost(t, newFakeRun())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Untitled list.mvd"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := sized(New(host, config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m, _ = choose(m, entryStart)
	if got := filepath.Base(m.session.Path()); got != "Untitled list 2.mvd" {
		t.Errorf("the draft is %q", got)
	}
}

func TestWithoutAFolderOfListsTheListIsAskedToBeSavedAsBefore(t *testing.T) {
	m := sized(New(Host{Start: func(config.Config, []string, []engine.PlannedEntry, bool) (ListRun, error) { return newFakeRun(), nil }}, config.Config{OutputDir: t.TempDir()}, downloadlist.New([]string{"https://youtu.be/x"})))
	m, _ = choose(m, entryStart)
	if m.asking != askSaveThenReview {
		t.Errorf("asking %v, want the question where to save", m.asking)
	}
}
