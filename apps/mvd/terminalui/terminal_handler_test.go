package terminalui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	ttygo "github.com/meta-tui/treactui/packages/tty-go"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/sourcelist"
	"youtube-downloader/libs/mvd-core/tui"
)

// noRun is a starter for tests that never start a run.
func noRun(config.Config, []string) (tui.Run, error) { return nil, errors.New("no run in this test") }

func testFiles(t *testing.T, urls ...string) (string, Files) {
	t.Helper()
	dir := t.TempDir()
	files := Files{Config: filepath.Join(dir, "config.conf"), List: filepath.Join(dir, "list.txt")}
	if err := config.Save(config.Default(dir, dir), files.Config); err != nil {
		t.Fatal(err)
	}
	if err := sourcelist.Save(files.List, urls); err != nil {
		t.Fatal(err)
	}
	return dir, files
}

func TestTheAppStartsOnTheSavedList(t *testing.T) {
	dir, files := testFiles(t, "https://a", "https://b")
	model := NewModelFactory(dir, files, Host{Start: noRun}, t.Logf)()

	snapshot := model.(accessibleApp).Accessible()
	if snapshot.Title != "MVD · Download list (2 items)" {
		t.Errorf("title: %q", snapshot.Title)
	}
}

func TestTheAppOpensThePreferencesOnTheFirstRun(t *testing.T) {
	dir := t.TempDir()
	files := Files{Config: filepath.Join(dir, "config.conf"), List: filepath.Join(dir, "list.txt")}
	model := NewModelFactory(dir, files, Host{Start: noRun}, t.Logf)()

	if title := model.(accessibleApp).Accessible().Title; title != "MVD · Preferences" {
		t.Errorf("the first run should open the preferences, got %q", title)
	}
}

// fakeEvents is the page's side of the event bus: it plays the page and keeps what Go sent.
type fakeEvents struct {
	mu       sync.Mutex
	handlers map[string]func(string)
	down     []string
}

func (e *fakeEvents) On(name string, handler func(string)) func() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handlers == nil {
		e.handlers = map[string]func(string){}
	}
	e.handlers[name] = handler
	return func() {}
}

func (e *fakeEvents) Emit(name, data string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if name == "treactui:down" {
		e.down = append(e.down, data)
	}
}

func (e *fakeEvents) page(data string) {
	e.mu.Lock()
	handler := e.handlers["treactui:up"]
	e.mu.Unlock()
	handler(data)
}

func (e *fakeEvents) seen() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return strings.Join(e.down, "")
}

func TestAPageThatConnectsIsGreetedAndShownTheApp(t *testing.T) {
	dir, files := testFiles(t, "https://a")
	events := &fakeEvents{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stop := ttygo.BindShared(ctx, NewShared(NewModelFactory(dir, files, Host{Start: noRun}, t.Logf)), events, ttygo.BindOptions{})
	defer stop()

	events.page(`{"c":"page","n":0,"t":"open"}`)

	for !strings.Contains(events.seen(), "Download list (1 item)") {
		if ctx.Err() != nil {
			t.Fatalf("never saw the list; got %q", events.seen())
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !strings.Contains(events.seen(), `type\":\"hello`) {
		t.Errorf("the first message should be a hello: %q", events.seen())
	}
}
