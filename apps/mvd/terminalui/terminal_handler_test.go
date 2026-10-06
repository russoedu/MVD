package terminalui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"net/http/httptest"

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

func TestAWindowThatConnectsIsGreetedAndShownTheApp(t *testing.T) {
	dir, files := testFiles(t, "https://a")
	server := httptest.NewServer(NewHandler(NewModelFactory(dir, files, Host{Start: noRun}, t.Logf), nil))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.CloseNow() }()

	seen := ""
	for !strings.Contains(seen, "Download list (1 item)") {
		_, frame, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("never saw the list; got %q (%v)", seen, err)
		}
		seen += string(frame)
	}
	if !strings.Contains(seen, `"type":"hello"`) {
		t.Errorf("the first frame should be a hello: %q", seen)
	}
}

func TestAPageFromAnotherSiteIsRefused(t *testing.T) {
	dir, files := testFiles(t)
	server := httptest.NewServer(NewHandler(NewModelFactory(dir, files, Host{Start: noRun}, t.Logf), nil))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{
		HTTPHeader: map[string][]string{"Origin": {"https://evil.example"}},
	})
	if err == nil {
		t.Error("a page from another site must not be able to open the app")
	}
}
