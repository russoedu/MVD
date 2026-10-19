package workspace

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/downloadlist"
	"youtube-downloader/libs/mvd-core/menubar"
)

func sized(m Model) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return next.(Model)
}

func press(m Model, key string) (Model, tea.Cmd) {
	var msg tea.KeyPressMsg
	if len([]rune(key)) == 1 {
		msg = tea.KeyPressMsg{Text: key, Code: []rune(key)[0]}
	} else {
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	}
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

func TestSavingAsWritesTheListAndKeepsItsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mine")
	m := sized(New(Host{ListDir: t.TempDir()}, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	next, _ := m.Update(menubar.SelectedMsg{Name: entrySaveAs})
	m = next.(Model)
	if m.asking != askSaveAs {
		t.Fatalf("asking %v", m.asking)
	}
	m.input.SetValue(path)
	m, _ = press(m, "enter")
	if got := m.session.Path(); got != path+".mvd" {
		t.Errorf("path %q", got)
	}
	if !strings.HasPrefix(m.notice, "Saved") {
		t.Errorf("notice %q", m.notice)
	}
}

func TestStartingWithNothingToStartWithSaysSo(t *testing.T) {
	host := Host{ListDir: t.TempDir()}
	m := sized(New(host, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	m, _ = press(m, "s")
	if m.notice != "Downloading is not available here" {
		t.Errorf("notice %q", m.notice)
	}
}

func TestAnEmptyListDoesNotStart(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, nil))
	m, _ = press(m, "s")
	if !strings.Contains(m.notice, "empty") {
		t.Errorf("notice %q", m.notice)
	}
}

func TestTheScreenFitsTheWindow(t *testing.T) {
	m := sized(New(Host{}, config.Config{}, downloadlist.New([]string{"https://youtu.be/x"})))
	lines := strings.Split(m.render(), "\n")
	if len(lines) != 30 {
		t.Errorf("%d rows", len(lines))
	}
	if !strings.Contains(lines[0], "MVD") || !strings.Contains(lines[0], "File") {
		t.Errorf("bar %q", lines[0])
	}
}
