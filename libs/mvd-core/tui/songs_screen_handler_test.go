package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
)

func typeText(m tea.Model, text string) tea.Model {
	m, _ = m.Update(tea.PasteMsg{Content: text})
	return m
}

func setupWithList(t *testing.T) (tea.Model, string) {
	t.Helper()
	dir := t.TempDir()
	var m tea.Model = newSetupModel(SetupInput{Cfg: config.Default(dir, dir), URLs: []string{"https://a"}, CfgPath: dir + "/config.conf", ListPath: dir + "/list.txt"})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return m, dir
}

func TestSongsScreenAddsAListFileToTheDownloadList(t *testing.T) {
	m, dir := setupWithList(t)

	m, _ = m.Update(key("ctrl+o"))
	if m.(setupModel).screen != screenSongs {
		t.Fatalf("ctrl+o should open the songs screen, on screen %d", m.(setupModel).screen)
	}
	m = typeText(m, "# Road trip\nATB - Killer\nSeal - Kiss From a Rose\nnot a song\n")
	if view := m.View().Content; !strings.Contains(view, "2 songs, 1 line not understood") {
		t.Errorf("the screen should say how the text reads:\n%s", view)
	}

	m, _ = m.Update(key("ctrl+s"))
	setup := m.(setupModel)
	want := filepath.Join(dir, "songs", "Road trip.txt")
	if setup.screen != screenList || setup.list.urls()[len(setup.list.urls())-1] != want {
		t.Fatalf("the list should end with %s, got screen %d and %v", want, setup.screen, setup.list.urls())
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("the list file should exist: %v", err)
	}
	if saved, _ := os.ReadFile(filepath.Join(dir, "list.txt")); !strings.Contains(string(saved), want) {
		t.Errorf("the download list file should name the songs file, got %q", saved)
	}
}

func TestSongsScreenRefusesATextWithoutSongsAndCancels(t *testing.T) {
	m, dir := setupWithList(t)
	m, _ = m.Update(key("ctrl+o"))
	m = typeText(m, "just words")

	m, _ = m.Update(key("ctrl+s"))
	if m.(setupModel).screen != screenSongs || !strings.Contains(m.View().Content, "Artist - Title") {
		t.Errorf("a text without songs stays on the screen with a hint:\n%s", m.View().Content)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "songs")); len(entries) != 0 {
		t.Errorf("nothing should be stored, found %d files", len(entries))
	}

	m, _ = m.Update(key("esc"))
	if setup := m.(setupModel); setup.screen != screenList || len(setup.list.urls()) != 1 {
		t.Errorf("esc goes back to the unchanged list, got screen %d and %v", setup.screen, setup.list.urls())
	}
}

func TestSongsScreenIsClickableAndDescribed(t *testing.T) {
	m, _ := setupWithList(t)
	m, _ = m.Update(click(16, 29)) // "songs" in the key bar
	if m.(setupModel).screen != screenSongs {
		t.Fatalf("clicking songs should open the screen, on screen %d", m.(setupModel).screen)
	}
	out := m.(outliner).Outline()
	if out.Title != "MVD · Add songs" || !out.HasText || len(out.Keys) != 2 {
		t.Errorf("unexpected outline %+v", out)
	}
}

func TestListWithURLDoesNotRepeatALine(t *testing.T) {
	m := newListModel([]string{"https://a"}).withURL("x.txt").withURL("x.txt")
	if got := m.urls(); len(got) != 2 || got[1] != "x.txt" {
		t.Errorf("got %v", got)
	}
}
